package session_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/oullin/alloy/pkg/hub/session"
	"github.com/oullin/alloy/pkg/hub/session/handlers"
)

// writeCountingHandler wraps ArrayHandler and counts store writes so a
// middleware-level test can assert a read-only request touches the backend
// zero times.
type writeCountingHandler struct {
	*handlers.ArrayHandler
	writes atomic.Int64
}

func (h *writeCountingHandler) Write(ctx context.Context, id, data string) error {
	h.writes.Add(1)

	return h.ArrayHandler.Write(ctx, id, data)
}

func TestStartSessionReadOnlyRequestIssuesNoStoreWrite(t *testing.T) {
	h := &writeCountingHandler{ArrayHandler: handlers.NewArrayHandler()}
	// A large ActivityRefresh keeps the sliding-expiry touch out of the way so
	// the second, read-only request exercises the pure skip-when-clean path.
	mw := session.StartSession(h, session.StartSessionConfig{
		CookieName:      "sess",
		GCProbability:   0,
		ActivityRefresh: time.Hour,
	})

	readOnly := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if store, ok := session.FromContext(r.Context()); ok {
			_ = store.Get("anything", nil) // read only
		}

		w.WriteHeader(http.StatusOK)
	})

	// First request establishes and persists a new session.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(readOnly).ServeHTTP(rr, req)

	var id string

	for _, c := range rr.Result().Cookies() {
		if c.Name == "sess" {
			id = c.Value
		}
	}

	if id == "" {
		t.Fatal("no session cookie on first request")
	}

	if got := h.writes.Load(); got != 1 {
		t.Fatalf("expected one write for the initial session, got %d", got)
	}

	// Second request reads the existing session and mutates nothing.
	h.writes.Store(0)

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: "sess", Value: id})
	mw(readOnly).ServeHTTP(rr2, req2)

	if got := h.writes.Load(); got != 0 {
		t.Errorf("a read-only request must issue zero store writes, got %d", got)
	}
}

func TestStartSessionWritesCookie(t *testing.T) {
	h := handlers.NewArrayHandler()
	mw := session.StartSession(h, session.StartSessionConfig{
		CookieName:    "sess",
		GCProbability: 0,
	})

	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(inner).ServeHTTP(rr, req)

	if !called {
		t.Fatal("inner handler was not called")
	}

	cookies := rr.Result().Cookies()

	var found bool

	for _, c := range cookies {
		if c.Name == "sess" {
			found = true
		}
	}

	if !found {
		t.Error("expected session cookie to be set")
	}
}

func TestStartSessionExposesStoreViaContext(t *testing.T) {
	h := handlers.NewArrayHandler()
	mw := session.StartSession(h, session.StartSessionConfig{
		CookieName:    "sess",
		GCProbability: 0,
	})

	var (
		got   *session.Store
		found bool
	)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, found = session.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(inner).ServeHTTP(rr, req)

	if !found {
		t.Fatal("expected store to be present in request context")
	}

	if got == nil {
		t.Fatal("expected non-nil store from context")
	}

	// The store exposed to the handler must be the one the middleware started
	// for this request, identified by the ID written to the response cookie.
	var cookieID string

	for _, c := range rr.Result().Cookies() {
		if c.Name == "sess" {
			cookieID = c.Value
		}
	}

	if got.GetID() != cookieID {
		t.Errorf("context store ID %q does not match cookie ID %q", got.GetID(), cookieID)
	}
}

func TestStartSessionMergesPartialConfigWithDefaults(t *testing.T) {
	h := handlers.NewArrayHandler()
	mw := session.StartSession(h, session.StartSessionConfig{
		Lifetime: time.Minute,
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	cookies := rr.Result().Cookies()

	if len(cookies) != 1 {
		t.Fatalf("expected one session cookie, got %d", len(cookies))
	}

	if cookies[0].Name != "session" {
		t.Fatalf("expected default cookie name session, got %q", cookies[0].Name)
	}

	if cookies[0].MaxAge != 60 {
		t.Fatalf("expected caller lifetime to be preserved as max-age 60, got %d", cookies[0].MaxAge)
	}
}

func TestStartSessionReusesID(t *testing.T) {
	h := handlers.NewArrayHandler()
	mw := session.StartSession(h, session.StartSessionConfig{
		CookieName:    "sess",
		GCProbability: 0,
	})

	var firstID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First request: get the session ID.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(inner).ServeHTTP(rr, req)

	for _, c := range rr.Result().Cookies() {
		if c.Name == "sess" {
			firstID = c.Value
		}
	}

	if firstID == "" {
		t.Fatal("no session cookie on first request")
	}

	// Second request with cookie: ID should be reused.
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: "sess", Value: firstID})
	mw(inner).ServeHTTP(rr2, req2)

	var secondID string

	for _, c := range rr2.Result().Cookies() {
		if c.Name == "sess" {
			secondID = c.Value
		}
	}

	if secondID != firstID {
		t.Errorf("session ID changed: got %q, want %q", secondID, firstID)
	}
}

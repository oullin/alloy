package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oullin/alloy/pkg/hub/session"
	"github.com/oullin/alloy/pkg/hub/session/handlers"
)

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

func TestStartSessionSecureTriState(t *testing.T) {
	cases := []struct {
		name   string
		secure *bool
		want   bool
	}{
		{name: "nil defaults to secure", secure: nil, want: true},
		{name: "explicit true honored", secure: session.BoolPtr(true), want: true},
		{name: "explicit false honored (dev opt-out)", secure: session.BoolPtr(false), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := handlers.NewArrayHandler()
			mw := session.StartSession(h, session.StartSessionConfig{
				CookieName:    "sess",
				GCProbability: 0,
				Secure:        tc.secure,
			})

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(rr, req)

			var found bool

			for _, c := range rr.Result().Cookies() {
				if c.Name == "sess" {
					found = true

					if c.Secure != tc.want {
						t.Fatalf("cookie Secure = %v, want %v", c.Secure, tc.want)
					}
				}
			}

			if !found {
				t.Fatal("expected session cookie to be set")
			}
		})
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

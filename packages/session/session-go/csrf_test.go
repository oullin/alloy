package session_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oullin/alloy/session"
)

type csrfStore struct{ token string }

func (s csrfStore) Token() string { return s.token }

func TestVerifyCSRFTokenAllowsReadRequests(t *testing.T) {
	called := false
	mw := session.VerifyCSRFToken(csrfStore{token: "expected"}, session.VerifyCSRFConfig{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	})).ServeHTTP(rec, req)

	if !called {
		t.Fatal("read request should pass without token")
	}
}

func TestVerifyCSRFTokenRejectsUnsafeRequestWithoutToken(t *testing.T) {
	mw := session.VerifyCSRFToken(csrfStore{token: "expected"}, session.VerifyCSRFConfig{})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler should not be called")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestVerifyCSRFTokenAcceptsHeaderToken(t *testing.T) {
	called := false
	mw := session.VerifyCSRFToken(csrfStore{token: "expected"}, session.VerifyCSRFConfig{})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-CSRF-Token", "expected")
	rec := httptest.NewRecorder()

	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	})).ServeHTTP(rec, req)

	if !called {
		t.Fatal("handler should be called with matching header token")
	}
}

func TestVerifyCSRFTokenAcceptsFormToken(t *testing.T) {
	called := false
	mw := session.VerifyCSRFToken(csrfStore{token: "expected"}, session.VerifyCSRFConfig{})
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("_token=expected"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	})).ServeHTTP(rec, req)

	if !called {
		t.Fatal("handler should be called with matching form token")
	}
}

func TestVerifyCSRFTokenDoesNotAcceptCookieTokenAlone(t *testing.T) {
	mw := session.VerifyCSRFToken(csrfStore{token: "expected"}, session.VerifyCSRFConfig{})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "XSRF-TOKEN", Value: "expected"})
	rec := httptest.NewRecorder()

	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler should not be called when token only came from a cookie")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestVerifyCSRFTokenHonorsExceptPaths(t *testing.T) {
	called := false
	mw := session.VerifyCSRFToken(csrfStore{token: "expected"}, session.VerifyCSRFConfig{
		Except: []string{"webhooks/*"},
	})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", nil)
	rec := httptest.NewRecorder()

	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	})).ServeHTTP(rec, req)

	if !called {
		t.Fatal("excepted path should pass without token")
	}
}

func TestVerifyCSRFTokenRejectsOriginMismatch(t *testing.T) {
	mw := session.VerifyCSRFToken(csrfStore{token: "expected"}, session.VerifyCSRFConfig{
		VerifyOrigin: true,
	})
	req := httptest.NewRequest(http.MethodPost, "https://app.test/profile", nil)
	req.Host = "app.test"
	req.Header.Set("Origin", "https://evil.test")
	req.Header.Set("X-CSRF-Token", "expected")
	rec := httptest.NewRecorder()

	mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler should not be called for origin mismatch")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

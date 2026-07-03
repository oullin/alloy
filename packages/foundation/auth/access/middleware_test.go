package access_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oullin/alloy/packages/foundation/auth/access"
	cauth "github.com/oullin/alloy/packages/foundation/contracts/auth"
)

type assertError string

func TestAuthorizeMiddlewareAllowsAuthorizedRequest(t *testing.T) {
	gate := access.New(userResolver(&testUser{id: "1"}))
	gate.Define("view", func(_ context.Context, _ cauth.User, _ any) (bool, error) {
		return true, nil
	})

	called := false
	handler := access.AuthorizeMiddleware(gate, "view", nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	if !called {
		t.Fatal("expected next handler to run")
	}
}

func TestAuthorizeMiddlewareRendersDeniedJSON(t *testing.T) {
	gate := access.New(userResolver(&testUser{id: "1"}))
	gate.Define("delete", func(_ context.Context, _ cauth.User, _ any) (bool, error) {
		return false, nil
	})

	handler := access.AuthorizeMiddleware(gate, "delete", nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not run")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content type = %q", rec.Header().Get("Content-Type"))
	}
}

func TestAuthorizeResolvedMiddlewareRendersNotFoundWhenModelCannotResolve(t *testing.T) {
	gate := access.New(userResolver(&testUser{id: "1"}))
	handler := access.AuthorizeResolvedMiddleware(gate, "view", func(*http.Request) (any, error) {
		return nil, assertError("missing")
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not run")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAuthorizeMiddlewareAppliesGateBeforeHooks(t *testing.T) {
	gate := access.New(userResolver(&testUser{id: "1"}))
	gate.Define("view", func(_ context.Context, _ cauth.User, _ any) (bool, error) {
		return false, nil
	})
	gate.Before(func(_ context.Context, _ cauth.User, ability string, _ any) (bool, bool) {
		if ability == "hidden" {
			return false, true
		}

		return false, false
	})

	handler := access.AuthorizeMiddleware(gate, "hidden", nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not run")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func (e assertError) Error() string { return string(e) }

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"alloy.dev/foundation/httpx/middleware"
)

func TestRecoveryPanicWritesInternalServerError(t *testing.T) {
	t.Parallel()

	handler := middleware.NewHandleRecovery().Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	if rec.Body.String() != `{"error":"internal server error"}` {
		t.Fatalf("expected JSON error body, got %s", rec.Body.String())
	}
}

func TestRecoveryRepanicsErrAbortHandler(t *testing.T) {
	t.Parallel()

	handler := middleware.NewHandleRecovery().Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		if v := recover(); v != http.ErrAbortHandler {
			t.Fatalf("expected ErrAbortHandler panic, got %v", v)
		}
	}()

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestRecoveryNormalHandlerUntouched(t *testing.T) {
	t.Parallel()

	handler := middleware.NewHandleRecovery().Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "ok")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	if rec.Header().Get("X-Test") != "ok" {
		t.Fatalf("expected header to pass through, got %s", rec.Header().Get("X-Test"))
	}

	if rec.Body.String() != "created" {
		t.Fatalf("expected body to pass through, got %s", rec.Body.String())
	}
}

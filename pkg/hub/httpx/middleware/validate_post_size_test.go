package middleware_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oullin/alloy/pkg/hub/httpx/middleware"
)

func TestValidatePostSizeAllowed(t *testing.T) {
	t.Parallel()

	handler := middleware.NewValidatePostSize(1024).Wrap(okHandler)

	body := strings.NewReader("small body")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Length", "10")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestValidatePostSizeRejected(t *testing.T) {
	t.Parallel()

	handler := middleware.NewValidatePostSize(5).Wrap(okHandler)

	body := strings.NewReader("this body is too large")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

// chunkedBodyHandler reads the request body and reports 413 when the wrapped
// reader trips its MaxBytesReader limit, echoing the observed ContentLength so
// the test can prove the request arrived chunked (ContentLength == -1).
func chunkedBodyHandler(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength != -1 {
		w.Header().Set("X-Content-Length", "declared")
	}

	if _, err := io.ReadAll(r.Body); err != nil {
		var mbe *http.MaxBytesError

		if errors.As(err, &mbe) {
			http.Error(w, "Post body too large", http.StatusRequestEntityTooLarge)

			return
		}

		http.Error(w, "read error", http.StatusBadRequest)

		return
	}

	w.WriteHeader(http.StatusOK)
}

func TestValidatePostSizeChunkedOversizeRejectedOnRead(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		middleware.NewValidatePostSize(10).Wrap(http.HandlerFunc(chunkedBodyHandler)),
	)
	defer srv.Close()

	// io.NopCloser hides the concrete reader type, so the client cannot compute
	// a Content-Length and falls back to Transfer-Encoding: chunked.
	body := io.NopCloser(strings.NewReader(strings.Repeat("a", 100)))
	req, err := http.NewRequest(http.MethodPost, srv.URL, body)

	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("Do: %v", err)
	}

	defer resp.Body.Close()

	if got := resp.Header.Get("X-Content-Length"); got == "declared" {
		t.Fatal("expected chunked request (ContentLength == -1), got a declared length")
	}

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversize chunked body, got %d", resp.StatusCode)
	}
}

func TestValidatePostSizeChunkedWithinLimitPasses(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		middleware.NewValidatePostSize(100).Wrap(http.HandlerFunc(chunkedBodyHandler)),
	)
	defer srv.Close()

	body := io.NopCloser(strings.NewReader(strings.Repeat("a", 10)))
	req, err := http.NewRequest(http.MethodPost, srv.URL, body)

	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("Do: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for within-limit chunked body, got %d", resp.StatusCode)
	}
}

func TestValidatePostSizeFromContentLengthHeader(t *testing.T) {
	t.Parallel()

	handler := middleware.NewValidatePostSize(100).Wrap(okHandler)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.ContentLength = 0
	req.Header.Set("Content-Length", "200")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

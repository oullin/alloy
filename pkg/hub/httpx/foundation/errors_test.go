package foundation_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/oullin/alloy/pkg/hub/httpx/foundation"
)

func TestHttpResponseErrorMessage(t *testing.T) {
	t.Parallel()

	err := foundation.NewHttpResponseError(http.StatusNotFound, "Not Found")

	expected := "httpx: HTTP 404: Not Found"

	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestNewHttpResponseErrorDefaults(t *testing.T) {
	t.Parallel()

	err := foundation.NewHttpResponseError(http.StatusBadRequest, "bad")

	if err.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", err.StatusCode)
	}

	if err.Message != "bad" {
		t.Fatalf("expected bad, got %s", err.Message)
	}

	if err.Headers == nil {
		t.Fatal("expected initialized headers")
	}
}

func TestThrottleRequestsError(t *testing.T) {
	t.Parallel()

	err := foundation.NewThrottleRequestsError(60)

	if err.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", err.StatusCode)
	}

	if err.RetryAfter != 60 {
		t.Fatalf("expected RetryAfter 60, got %d", err.RetryAfter)
	}

	if err.Headers.Get("Retry-After") != "60" {
		t.Fatalf("expected Retry-After header, got %q", err.Headers.Get("Retry-After"))
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{"ErrPostTooLarge", foundation.ErrPostTooLarge},
		{"ErrMalformedURL", foundation.ErrMalformedURL},
		{"ErrOriginMismatch", foundation.ErrOriginMismatch},
		{"ErrThrottle", foundation.ErrThrottle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.err == nil {
				t.Fatal("sentinel error should not be nil")
			}

			if tt.err.Error() == "" {
				t.Fatal("sentinel error message should not be empty")
			}
		})
	}
}

func TestSentinelErrorsAreDistinct(t *testing.T) {
	t.Parallel()

	if errors.Is(foundation.ErrPostTooLarge, foundation.ErrMalformedURL) {
		t.Fatal("ErrPostTooLarge and ErrMalformedURL should be distinct")
	}

	if errors.Is(foundation.ErrOriginMismatch, foundation.ErrThrottle) {
		t.Fatal("ErrOriginMismatch and ErrThrottle should be distinct")
	}
}

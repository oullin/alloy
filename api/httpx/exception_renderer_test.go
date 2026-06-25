package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExceptionRendererRendersValidationErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	renderer := ExceptionRenderer{}

	renderer.Render(rec, &ValidationError{
		Errors: map[string][]string{"email": {"required"}},
	})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}

	if !strings.Contains(rec.Body.String(), `"errors"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestExceptionRendererRendersHttpResponseErrorHeadersAndPayload(t *testing.T) {
	rec := httptest.NewRecorder()
	err := NewHttpResponseError(http.StatusTooManyRequests, "Too many attempts")
	err.Headers.Set("Retry-After", "60")
	err.Response = map[string]any{"code": "lockout"}

	ExceptionRenderer{}.Render(rec, err)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	if rec.Header().Get("Retry-After") != "60" {
		t.Fatalf("retry-after = %q", rec.Header().Get("Retry-After"))
	}

	if !strings.Contains(rec.Body.String(), `"code":"lockout"`) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestExceptionRendererHidesUnexpectedErrorsUnlessDebug(t *testing.T) {
	rec := httptest.NewRecorder()

	ExceptionRenderer{}.Render(rec, errors.New("database password leaked"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	if strings.Contains(rec.Body.String(), "database password leaked") {
		t.Fatalf("body leaked error: %s", rec.Body.String())
	}

	debug := httptest.NewRecorder()
	ExceptionRenderer{Debug: true}.Render(debug, errors.New("diagnostic"))

	if !strings.Contains(debug.Body.String(), "diagnostic") {
		t.Fatalf("debug body = %s", debug.Body.String())
	}
}

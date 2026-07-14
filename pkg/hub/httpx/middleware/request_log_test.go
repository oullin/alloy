package middleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/oullin/alloy/pkg/hub/httpx/middleware"
)

type slogCapture struct {
	mu      sync.Mutex
	entries []slog.Record
}

var slogMu sync.Mutex

func TestRequestLogCapturesStatusAndBytes(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.Handler
		wantStatus int64
		wantBytes  int64
	}{
		{
			name: "explicit status",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("created"))
			}),
			wantStatus: http.StatusCreated,
			wantBytes:  int64(len("created")),
		},
		{
			name: "implicit status",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("ok"))
			}),
			wantStatus: http.StatusOK,
			wantBytes:  int64(len("ok")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := useSlogCapture(t)
			handler := middleware.NewHandleRequestLog(middleware.RequestLogOptions{}).Wrap(tt.handler)

			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items", nil))

			records := capture.records("http: request")

			if len(records) != 1 {
				t.Fatalf("expected one log record, got %d", len(records))
			}

			attrs := recordAttrs(records[0])

			if attrs["status"].Int64() != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, attrs["status"].Int64())
			}

			if attrs["bytes"].Int64() != tt.wantBytes {
				t.Fatalf("expected bytes %d, got %d", tt.wantBytes, attrs["bytes"].Int64())
			}

			if attrs["method"].String() != http.MethodGet {
				t.Fatalf("expected method GET, got %s", attrs["method"].String())
			}

			if attrs["path"].String() != "/items" {
				t.Fatalf("expected path /items, got %s", attrs["path"].String())
			}
		})
	}
}

func TestRequestLogSkipPathDoesNotLog(t *testing.T) {
	capture := useSlogCapture(t)
	called := false
	handler := middleware.NewHandleRequestLog(middleware.RequestLogOptions{
		SkipPaths: []string{"/health"},
	}).Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))

	if !called {
		t.Fatal("expected skipped request to pass through")
	}

	if got := len(capture.records("http: request")); got != 0 {
		t.Fatalf("expected no log records, got %d", got)
	}
}

func TestRequestLogFlusherPassthrough(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	var sawFlusher bool
	handler := middleware.NewHandleRequestLog(middleware.RequestLogOptions{}).Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		sawFlusher = ok

		if ok {
			flusher.Flush()
		}
	}))

	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !sawFlusher {
		t.Fatal("expected wrapped response writer to implement http.Flusher")
	}

	if !rec.Flushed {
		t.Fatal("expected underlying response writer to be flushed")
	}
}

func useSlogCapture(t *testing.T) *slogCapture {
	t.Helper()

	slogMu.Lock()

	capture := &slogCapture{}
	previous := slog.Default()
	slog.SetDefault(slog.New(capture))

	t.Cleanup(func() {
		slog.SetDefault(previous)
		slogMu.Unlock()
	})

	return capture
}

func (h *slogCapture) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *slogCapture) Handle(ctx context.Context, record slog.Record) error {
	h.mu.Lock()

	defer h.mu.Unlock()

	h.entries = append(h.entries, record.Clone())

	return nil
}

func (h *slogCapture) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *slogCapture) WithGroup(name string) slog.Handler {
	return h
}

func (h *slogCapture) records(message string) []slog.Record {
	h.mu.Lock()

	defer h.mu.Unlock()

	records := make([]slog.Record, 0, len(h.entries))

	for _, record := range h.entries {
		if record.Message == message {
			records = append(records, record)
		}
	}

	return records
}

func recordAttrs(record slog.Record) map[string]slog.Value {
	attrs := map[string]slog.Value{}

	record.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value

		return true
	})

	return attrs
}

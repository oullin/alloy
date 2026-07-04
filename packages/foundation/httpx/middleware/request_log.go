package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// RequestLogOptions configures request logging behaviour.
type RequestLogOptions struct {
	SkipPaths []string
}

// HandleRequestLog logs completed HTTP requests.
type HandleRequestLog struct {
	opts RequestLogOptions
}

// NewHandleRequestLog creates the request logging middleware.

// Wrap returns an http.Handler that logs requests.

type requestLogResponseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

type flushingRequestLogResponseWriter struct {
	*requestLogResponseWriter
	flusher http.Flusher
}

func NewHandleRequestLog(opts RequestLogOptions) *HandleRequestLog {
	return &HandleRequestLog{opts: opts}
}

func (m *HandleRequestLog) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.shouldSkip(r.URL.Path) {
			next.ServeHTTP(w, r)

			return
		}

		start := time.Now()
		rec := &requestLogResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		rw := http.ResponseWriter(rec)

		if flusher, ok := w.(http.Flusher); ok {
			rw = &flushingRequestLogResponseWriter{
				requestLogResponseWriter: rec,
				flusher:                  flusher,
			}
		}

		next.ServeHTTP(rw, r)

		slog.Info("http: request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start),
			"bytes", rec.bytes,
		)
	})
}

func (m *HandleRequestLog) shouldSkip(path string) bool {
	for _, skipPath := range m.opts.SkipPaths {
		if path == skipPath {
			return true
		}
	}

	return false
}

func (w *requestLogResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.status = http.StatusOK
		w.wroteHeader = true
	}

	n, err := w.ResponseWriter.Write(body)
	w.bytes += n

	return n, err
}

func (w *requestLogResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.status = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *requestLogResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *flushingRequestLogResponseWriter) Flush() {
	if !w.wroteHeader {
		w.status = http.StatusOK
		w.wroteHeader = true
	}

	w.flusher.Flush()
}

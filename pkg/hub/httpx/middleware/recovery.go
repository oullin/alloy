package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// HandleRecovery recovers panics from downstream handlers.
type HandleRecovery struct{}

// NewHandleRecovery creates the recovery middleware.

// Wrap returns an http.Handler that recovers panics.

type recoveryResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func NewHandleRecovery() *HandleRecovery {
	return &HandleRecovery{}
}

func (m *HandleRecovery) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &recoveryResponseWriter{ResponseWriter: w}

		defer func() {
			v := recover()

			if v == nil {
				return
			}

			if v == http.ErrAbortHandler {
				panic(v)
			}

			slog.Error("recovery: panic recovered", "error", v, "stack", string(debug.Stack()))

			if rec.wroteHeader {
				return
			}

			rec.Header().Set("Content-Type", "application/json")
			rec.WriteHeader(http.StatusInternalServerError)
			rec.Write([]byte(`{"error":"internal server error"}`))
		}()

		next.ServeHTTP(rec, r)
	})
}

// Unwrap exposes the underlying writer so http.ResponseController can reach
// Flusher/Hijacker implementations through the middleware chain.
func (w *recoveryResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *recoveryResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
	}

	return w.ResponseWriter.Write(body)
}

func (w *recoveryResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

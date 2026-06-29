package observability

import (
	"net/http"
	"time"

	cauth "alloy.dev/go/contracts/auth"
	clog "alloy.dev/go/contracts/auth/log"
)

// RequestIDResolver returns the request id for a request.
type RequestIDResolver func(r *http.Request) string

// Middleware enriches the request context and emits structured request logs.
func Middleware(logger clog.Sink, guard cauth.Guard, resolve RequestIDResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			userID := ""

			if guard != nil {
				if user, err := guard.User(r.Context()); err == nil && user != nil {
					userID = user.GetAuthIdentifier()
				}
			}

			requestID := ""

			if resolve != nil {
				requestID = resolve(r)
			}

			meta := FromHTTPRequest(r, requestID, userID)
			next.ServeHTTP(w, r.WithContext(WithRequestContext(r.Context(), meta)))

			if logger != nil {
				logger.Info("auth request handled", map[string]any{
					"request_id":  meta.RequestID,
					"user_id":     meta.UserID,
					"ip":          meta.IP,
					"user_agent":  meta.UserAgent,
					"method":      meta.Method,
					"path":        meta.Path,
					"duration_ms": time.Since(start).Milliseconds(),
				})
			}
		})
	}
}

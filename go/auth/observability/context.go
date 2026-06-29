package observability

import (
	"context"
	"net/http"
)

type contextKey string

// RequestContext carries stable request metadata through auth flows.
type RequestContext struct {
	RequestID string
	UserID    string
	IP        string
	UserAgent string
	Path      string
	Method    string
}

const requestContextKey contextKey = "auth_observability_request"

// WithRequestContext stores request metadata on ctx.
func WithRequestContext(ctx context.Context, meta RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey, meta)
}

// RequestContextFromContext returns request metadata from ctx.
func RequestContextFromContext(ctx context.Context) (RequestContext, bool) {
	meta, ok := ctx.Value(requestContextKey).(RequestContext)

	return meta, ok
}

// FromHTTPRequest creates request metadata from an HTTP request.
func FromHTTPRequest(r *http.Request, requestID, userID string) RequestContext {
	return RequestContext{
		RequestID: requestID,
		UserID:    userID,
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
		Path:      r.URL.Path,
		Method:    r.Method,
	}
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}

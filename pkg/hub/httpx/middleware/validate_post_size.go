package middleware

import (
	"net/http"
	"strconv"
)

// ValidatePostSize rejects requests whose Content-Length exceeds a maximum.
type ValidatePostSize struct {
	maxBytes int64
}

// NewValidatePostSize creates the middleware with the given byte limit.
func NewValidatePostSize(maxBytes int64) *ValidatePostSize {
	return &ValidatePostSize{maxBytes: maxBytes}
}

// Wrap returns an http.Handler that checks Content-Length before delegating.
func (m *ValidatePostSize) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > m.maxBytes {
			http.Error(w, "Post body too large", http.StatusRequestEntityTooLarge)

			return
		}

		// Also check the Content-Length header string for cases where
		// ContentLength might not be set but the header is present.
		if cl := r.Header.Get("Content-Length"); cl != "" {
			if size, err := strconv.ParseInt(cl, 10, 64); err == nil && size > m.maxBytes {
				http.Error(w, "Post body too large", http.StatusRequestEntityTooLarge)

				return
			}
		}

		// The header fast-paths above cannot catch Transfer-Encoding: chunked
		// bodies (ContentLength == -1, header stripped). Wrap the body so an
		// oversize undeclared/chunked stream fails on read: the handler sees a
		// *http.MaxBytesError and the response is committed as 413.
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, m.maxBytes)
		}

		next.ServeHTTP(w, r)
	})
}

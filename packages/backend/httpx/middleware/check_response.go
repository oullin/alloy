package middleware

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"
)

// CheckResponseForModifications handles conditional requests using ETag and
// If-Modified-Since headers. When the response has not been modified, a 304
// Not Modified is returned instead of the full body.
type CheckResponseForModifications struct{}

// NewCheckResponseForModifications creates the middleware.

// Wrap returns an http.Handler that checks ETags and modification times.

// Only apply to safe methods.

// Capture the response metadata and body without importing httptest in production code.

// Generate ETag if not present.

// Check If-None-Match.

// Check If-Modified-Since.

// Forward the full response.

type bufferedResponseWriter struct {
	HeaderMap http.Header
	Body      bytes.Buffer
	Code      int
}

func NewCheckResponseForModifications() *CheckResponseForModifications {
	return &CheckResponseForModifications{}
}

func (m *CheckResponseForModifications) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			next.ServeHTTP(w, r)

			return
		}

		rec := newBufferedResponseWriter()
		next.ServeHTTP(rec, r)

		body := rec.Body.Bytes()

		etag := rec.Header().Get("ETag")

		if etag == "" && len(body) > 0 {
			hash := sha256.Sum256(body)
			etag = fmt.Sprintf(`"%x"`, hash[:8])
			rec.Header().Set("ETag", etag)
		}

		if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" && etag != "" {
			if ifNoneMatch == etag {
				copyHeaders(w, rec)
				w.WriteHeader(http.StatusNotModified)

				return
			}
		}

		if ims := r.Header.Get("If-Modified-Since"); ims != "" {
			if lastMod := rec.Header().Get("Last-Modified"); lastMod != "" {
				imsTime, err1 := time.Parse(http.TimeFormat, ims)
				lmTime, err2 := time.Parse(http.TimeFormat, lastMod)

				if err1 == nil && err2 == nil && !lmTime.After(imsTime) {
					copyHeaders(w, rec)
					w.WriteHeader(http.StatusNotModified)

					return
				}
			}
		}

		copyHeaders(w, rec)
		w.WriteHeader(rec.Code)
		w.Write(body)
	})
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		HeaderMap: make(http.Header),
		Code:      http.StatusOK,
	}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.HeaderMap
}

func (w *bufferedResponseWriter) Write(body []byte) (int, error) {
	return w.Body.Write(body)
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
}

func copyHeaders(w http.ResponseWriter, rec *bufferedResponseWriter) {
	for k, vals := range rec.Header() {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
}

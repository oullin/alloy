package session

import (
	"crypto/subtle"
	"net/http"
	"strings"

	csession "github.com/oullin/alloy/packages/foundation/contracts/session"
)

// CSRFStore is the session surface required by VerifyCSRFToken.
type CSRFStore = csession.CSRFStore

// VerifyCSRFConfig configures CSRF request verification.
type VerifyCSRFConfig = csession.VerifyCSRFConfig

// VerifyCSRFToken rejects unsafe requests whose submitted token does not match
// the session token. It accepts tokens from a form field, X-CSRF-Token, or the
// X-XSRF-Token header used by SPA clients.
func VerifyCSRFToken(store CSRFStore, cfg VerifyCSRFConfig) func(http.Handler) http.Handler {
	cfg = mergeVerifyCSRFConfig(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if store == nil || isReading(r.Method) || pathIsExcepted(r.URL.Path, cfg.Except) {
				next.ServeHTTP(w, r)

				return
			}

			if cfg.VerifyOrigin && !hasValidOrigin(r) {
				http.Error(w, "CSRF origin mismatch", http.StatusForbidden)

				return
			}

			if !tokensEqual(store.Token(), tokenFromRequest(r, cfg)) {
				http.Error(w, "CSRF token mismatch", http.StatusForbidden)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func mergeVerifyCSRFConfig(cfg VerifyCSRFConfig) VerifyCSRFConfig {
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-CSRF-Token"
	}

	if cfg.FormField == "" {
		cfg.FormField = "_token"
	}

	return cfg
}

func isReading(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func tokenFromRequest(r *http.Request, cfg VerifyCSRFConfig) string {
	if token := r.Header.Get(cfg.HeaderName); token != "" {
		return token
	}

	if token := r.Header.Get("X-XSRF-Token"); token != "" {
		return token
	}

	if token := r.FormValue(cfg.FormField); token != "" {
		return token
	}

	return ""
}

func tokensEqual(expected, actual string) bool {
	if expected == "" || actual == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}

func pathIsExcepted(path string, except []string) bool {
	path = strings.TrimPrefix(path, "/")

	for _, pattern := range except {
		pattern = strings.Trim(strings.TrimPrefix(pattern, "/"), " ")

		if pattern == "" {
			continue
		}

		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")

			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				return true
			}

			continue
		}

		if path == pattern {
			return true
		}
	}

	return false
}

func hasValidOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	if origin == "" {
		origin = r.Header.Get("Referer")
	}

	if origin == "" {
		return true
	}

	return sameHost(origin, r.Host)
}

func sameHost(rawURL, host string) bool {
	rawURL = strings.TrimSpace(rawURL)

	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(rawURL, prefix) {
			withoutScheme := strings.TrimPrefix(rawURL, prefix)
			candidate := withoutScheme

			if slash := strings.Index(candidate, "/"); slash >= 0 {
				candidate = candidate[:slash]
			}

			return strings.EqualFold(candidate, host)
		}
	}

	return strings.EqualFold(rawURL, host)
}

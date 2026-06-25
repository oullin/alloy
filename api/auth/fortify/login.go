package fortify

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	cauth "github.com/oullin/alloy/auth/contracts/auth"
)

// LoginConfig controls how credentials are read from a headless login request.
type LoginConfig struct {
	UsernameField string
	PasswordField string
	Limiter       LoginLimiter
	LimitKey      func(r *http.Request, username string) string
}

// NewLoginHandler authenticates a user and lets the guard own session migration.
func NewLoginHandler(guard cauth.StatefulGuard, cfg LoginConfig) http.HandlerFunc {
	usernameField := cfg.UsernameField

	if usernameField == "" {
		usernameField = "email"
	}

	passwordField := cfg.PasswordField

	if passwordField == "" {
		passwordField = "password"
	}

	return func(w http.ResponseWriter, r *http.Request) {
		input, err := readInput(r)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")

			return
		}

		username := stringInput(input, usernameField)
		password := stringInput(input, passwordField)

		if username == "" {
			writeValidation(w, usernameField, "required")

			return
		}

		if password == "" {
			writeValidation(w, passwordField, "required")

			return
		}

		limitKey := loginLimitKey(r, username, cfg.LimitKey)

		if cfg.Limiter != nil && cfg.Limiter.TooManyAttempts(r.Context(), limitKey) {
			w.Header().Set("Retry-After", retryAfterSeconds(cfg.Limiter.AvailableIn(r.Context(), limitKey)))
			writeError(w, http.StatusTooManyRequests, "too many login attempts")

			return
		}

		credentials := map[string]string{usernameField: username, passwordField: password}

		if !guard.Attempt(r.Context(), credentials, boolInput(input, "remember")) {
			if cfg.Limiter != nil {
				_ = cfg.Limiter.Hit(r.Context(), limitKey)
			}

			writeValidation(w, usernameField, "invalid credentials")

			return
		}

		if cfg.Limiter != nil {
			_ = cfg.Limiter.Clear(r.Context(), limitKey)
		}

		writeOK(w, "authenticated")
	}
}

// NewLogoutHandler logs out the current session.
func NewLogoutHandler(guard cauth.StatefulGuard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := guard.Logout(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, "logout failed")

			return
		}

		writeNoContent(w)
	}
}

func loginLimitKey(r *http.Request, username string, custom func(*http.Request, string) string) string {
	if custom != nil {
		return custom(r, username)
	}

	host, _, ok := strings.Cut(r.RemoteAddr, ":")

	if !ok {
		host = r.RemoteAddr
	}

	return strings.ToLower(username) + "|" + host
}

func retryAfterSeconds(duration time.Duration) string {
	seconds := int(duration.Seconds())

	if seconds < 1 {
		seconds = 1
	}

	return strconv.Itoa(seconds)
}

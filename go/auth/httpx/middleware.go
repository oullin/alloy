package httpx

import (
	"context"
	"net/http"
	"time"

	"alloy.dev/api/auth/security"
	cauth "alloy.dev/api/contracts/auth"
)

type contextKey string

// PasswordSession is the minimal session shape needed for password confirmation.
type PasswordSession interface {
	Get(key string, fallback any) any
}

const userContextKey contextKey = "auth_user"

// WithUser stores an authenticated user in the request context.
func WithUser(r *http.Request, user cauth.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userContextKey, user))
}

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(ctx context.Context) cauth.User {
	u, _ := ctx.Value(userContextKey).(cauth.User)

	return u
}

// EnsureAuthenticated rejects unauthenticated requests with 401.
func EnsureAuthenticated(guard cauth.Guard) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := guard.User(r.Context())

			if err != nil || user == nil {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(w, WithUser(r, user))
		})
	}
}

// RedirectIfAuthenticated redirects authenticated users to the given path.
func RedirectIfAuthenticated(guard cauth.Guard, redirectTo string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if guard.Check(r.Context()) {
				http.Redirect(w, r, redirectTo, http.StatusFound)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// EnsureEmailIsVerified rejects users whose email is not verified.
func EnsureEmailIsVerified(guard cauth.Guard) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := guard.User(r.Context())

			if err != nil || user == nil {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)

				return
			}

			if mv, ok := user.(cauth.MustVerifyEmail); ok && !mv.HasVerifiedEmail() {
				http.Error(w, "email not verified", http.StatusForbidden)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePassword redirects the user to the password confirmation page if they
// have not recently confirmed their password within the given timeout.
func RequirePassword(session PasswordSession, timeout time.Duration, confirmPath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const key = "auth.password_confirmed_at"
			confirmedAt := sessionTimestamp(session.Get(key, int64(0)))
			threshold := time.Now().Add(-timeout).Unix()

			if confirmedAt < threshold {
				http.Redirect(w, r, confirmPath, http.StatusFound)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func sessionTimestamp(value any) int64 {
	const maxInt64 = uint64(1<<63 - 1)

	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int16:
		return int64(v)
	case int8:
		return int64(v)
	case uint:
		if uint64(v) > maxInt64 {
			return 0
		}

		return int64(v)
	case uint64:
		if v > maxInt64 {
			return 0
		}

		return int64(v)
	case uint32:
		return int64(v)
	case uint16:
		return int64(v)
	case uint8:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	default:
		return 0
	}
}

// WithBasicAuth attempts HTTP Basic authentication on each request.
func WithBasicAuth(provider cauth.UserProvider, hasher cauth.PasswordHasher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()

			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)

				return
			}

			var (
				user  cauth.User
				match bool
				err   error
			)

			security.Timebox(75*time.Millisecond, func() {
				user, err = provider.RetrieveByCredentials(r.Context(), map[string]string{"email": username})

				passwordHash := ""

				if err == nil && user != nil {
					passwordHash = user.GetAuthPassword()
				}

				var checkErr error

				match, checkErr = hasher.Check(r.Context(), password, passwordHash)

				if err == nil {
					err = checkErr
				}
			})

			if err != nil || user == nil || !match {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(w, WithUser(r, user))
		})
	}
}

package tokens

import (
	"context"
	"net/http"
	"strings"

	authroot "github.com/oullin/alloy/pkg/hub/auth/httpx"
	cauth "github.com/oullin/alloy/pkg/hub/contracts/auth"
)

type contextKey string

const tokenContextKey contextKey = "auth_personal_access_token"

// CurrentToken returns the personal access token resolved for the request.
func CurrentToken(ctx context.Context) (Token, bool) {
	token, ok := ctx.Value(tokenContextKey).(Token)

	return token, ok
}

// WithToken stores a personal access token in context.
func WithToken(ctx context.Context, token Token) context.Context {
	return context.WithValue(ctx, tokenContextKey, token)
}

// AuthenticateBearerToken authenticates Authorization: Bearer id|secret tokens.
func AuthenticateBearerToken(repo Repository, users cauth.UserProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			plainText := bearerToken(r)

			if plainText == "" {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)

				return
			}

			token, err := FindByPlainTextToken(r.Context(), repo, plainText)

			if err != nil {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)

				return
			}

			user, err := users.RetrieveByID(r.Context(), token.UserID)

			if err != nil || user == nil {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)

				return
			}

			_ = repo.Touch(r.Context(), token.ID)
			ctx := WithToken(r.Context(), token)
			next.ServeHTTP(w, authroot.WithUser(r.WithContext(ctx), user))
		})
	}
}

// RequireAbility rejects requests whose current token lacks the ability.
func RequireAbility(ability string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := CurrentToken(r.Context())

			if !ok || token.Cant(ability) {
				http.Error(w, "forbidden", http.StatusForbidden)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
}

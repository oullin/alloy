package fortify

import (
	"context"
	"net/http"
	"time"

	"hara.sh/alloy/auth/passwords"
	cauth "hara.sh/alloy/contracts/auth"
)

// RegisterUser creates a user from a headless registration request.
type RegisterUser func(ctx context.Context, input RegisterInput) (cauth.User, error)

// ResetLinkSender sends password reset links without exposing the delivery
// mechanism to HTTP handlers.
type ResetLinkSender interface {
	SendResetLink(ctx context.Context, email string) error
}

// PasswordResetter validates reset credentials and applies the supplied reset
// callback when they are valid.
type PasswordResetter interface {
	Reset(ctx context.Context, credentials map[string]any, resetFn passwords.ResetCallback) error
}

// VerifyEmail verifies the current user's signed email-verification request.
type VerifyEmail func(ctx context.Context, r *http.Request, user cauth.MustVerifyEmail) error

// PasswordConfirmationSession stores password confirmation timestamps.
type PasswordConfirmationSession interface {
	Put(key string, value any)
}

// ProfileUpdater persists user profile changes.
type ProfileUpdater func(ctx context.Context, user cauth.User, input map[string]any) error

// PasswordUpdater persists a newly-hashed user password.
type PasswordUpdater func(ctx context.Context, user cauth.User, hashedPassword string) error

// PasswordSessionInvalidator revokes sessions or credentials after password changes.
type PasswordSessionInvalidator func(ctx context.Context, user cauth.User) error

// TwoFactorUpdater persists changes to a user's two-factor state.
type TwoFactorUpdater func(ctx context.Context, user cauth.TwoFactorUser) error

// CurrentSessionID resolves the active browser session ID for a request.
type CurrentSessionID func(r *http.Request) string

// PasskeySessionKey resolves the server-side WebAuthn ceremony key.
type PasskeySessionKey func(r *http.Request) string

// PasskeyUserResolver resolves a user by application user ID after passkey login.
type PasskeyUserResolver func(ctx context.Context, userID string) (cauth.User, error)

// PasskeyService is the passkey ceremony surface Fortify needs. Heavy WebAuthn
// implementations live outside the parent foundation module.
type PasskeyService interface {
	BeginRegistration(ctx context.Context, key string, user cauth.User) (any, error)
	FinishRegistration(ctx context.Context, key string, user cauth.User, r *http.Request) (any, error)
	BeginDiscoverableLogin(ctx context.Context, key string) (any, error)
	FinishPasskeyLogin(ctx context.Context, key string, r *http.Request, resolveUser PasskeyUserResolver) (cauth.User, any, error)
}

// LoginLimiter tracks failed login attempts and lockouts.
type LoginLimiter interface {
	TooManyAttempts(ctx context.Context, key string) bool
	Hit(ctx context.Context, key string) error
	Clear(ctx context.Context, key string) error
	AvailableIn(ctx context.Context, key string) time.Duration
}

const PasswordConfirmedAtKey = "auth.password_confirmed_at"

package auth

import (
	"context"
	"time"
)

// User is any entity that can be authenticated.
type User interface {
	GetAuthIdentifierName() string
	GetAuthIdentifier() string
	GetAuthPasswordName() string
	GetAuthPassword() string
	SetAuthPassword(password string)
	GetRememberToken() string
	SetRememberToken(token string)
	GetRememberTokenName() string
}

// BroadcastingUser is implemented by users that expose a stable
// authentication identifier for private broadcast channel authorization.
type BroadcastingUser interface {
	User
	GetAuthIdentifierForBroadcasting() string
}

// MustVerifyEmail is implemented by users that require email verification.
type MustVerifyEmail interface {
	HasVerifiedEmail() bool
	MarkEmailAsVerified(at time.Time)
	MarkEmailAsUnverified()
	GetEmailForVerification() string
}

// EmailVerificationNotificationSender is implemented by users that can send
// their own email verification notification.
type EmailVerificationNotificationSender interface {
	MustVerifyEmail
	SendEmailVerificationNotification(ctx context.Context)
}

// CanResetPassword is implemented by users that support password resets.
type CanResetPassword interface {
	GetEmailForPasswordReset() string
}

// ResettableUser is implemented by authenticated users that support
// password resets.
type ResettableUser interface {
	User
	CanResetPassword
}

// PasswordResetNotificationSender is implemented by users that can send their
// own password reset notification.
type PasswordResetNotificationSender interface {
	CanResetPassword
	SendPasswordResetNotification(ctx context.Context, token string)
}

// TwoFactorUser is implemented by users that support 2FA.
type TwoFactorUser interface {
	IsTwoFactorEnabled() bool
	SetTwoFactorEnabled(enabled bool)
	GetTwoFactorSecret() string
	SetTwoFactorSecret(secret string)
	GetTwoFactorRecoveryCodes() []string
	SetTwoFactorRecoveryCodes(codes []string)
	GetTwoFactorConfirmedAt() *time.Time
	SetTwoFactorConfirmedAt(at *time.Time)
}

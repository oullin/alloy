package listeners

import (
	"context"

	"github.com/oullin/alloy/pkg/hub/auth/events"
	cauth "github.com/oullin/alloy/pkg/hub/contracts/auth"
)

// EmailVerificationSender is kept for backward compatibility. New code should
// use contracts/auth.EmailVerificationNotificationSender.
type EmailVerificationSender = cauth.EmailVerificationNotificationSender

// SendEmailVerificationNotification sends an email verification notification
// when a new user registers, if the user implements EmailVerificationSender and has
// not yet verified their email.
type SendEmailVerificationNotification struct{}

// Handle processes a Registered event.
func (l *SendEmailVerificationNotification) Handle(ctx context.Context, event events.Registered) {
	mv, ok := event.User.(cauth.EmailVerificationNotificationSender)

	if !ok {
		return
	}

	if mv.HasVerifiedEmail() {
		return
	}

	mv.SendEmailVerificationNotification(ctx)
}

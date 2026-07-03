package notifications

import (
	"context"
	"errors"

	cauth "github.com/oullin/alloy/packages/foundation/contracts/auth"
)

// ResetURLBuilder builds a reset-password URL for a user and token.
type ResetURLBuilder func(ctx context.Context, user cauth.CanResetPassword, token string) (string, error)

// ResetPassword renders the password reset mail notification.
type ResetPassword struct {
	Token      string
	URLBuilder ResetURLBuilder
	Subject    string
}

// ToMail renders a reset password message.
func (n ResetPassword) ToMail(ctx context.Context, notifiable any) (MailMessage, error) {
	user, ok := notifiable.(cauth.CanResetPassword)

	if !ok {
		return MailMessage{}, errors.New("notifications: notifiable cannot reset password")
	}

	actionURL := ""

	if n.URLBuilder != nil {
		url, err := n.URLBuilder(ctx, user, n.Token)

		if err != nil {
			return MailMessage{}, err
		}

		actionURL = url
	}

	subject := n.Subject

	if subject == "" {
		subject = "Reset Password Notification"
	}

	return MailMessage{
		Subject:    subject,
		IntroLines: []string{"You are receiving this email because we received a password reset request for your account."},
		ActionText: "Reset Password",
		ActionURL:  actionURL,
		OutroLines: []string{"If you did not request a password reset, no further action is required."},
	}, nil
}

func (n ResetPassword) ShouldQueue() bool { return false }
func (n ResetPassword) QueueConnection() string {
	return ""
}
func (n ResetPassword) QueueName() string {
	return ""
}

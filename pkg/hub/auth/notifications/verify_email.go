package notifications

import (
	"context"
	"errors"

	cauth "hara.sh/alloy/contracts/auth"
)

// VerificationURLBuilder builds a signed verification URL for a user.
type VerificationURLBuilder func(ctx context.Context, user cauth.MustVerifyEmail) (string, error)

// VerifyEmail renders the email verification notification.
type VerifyEmail struct {
	URLBuilder VerificationURLBuilder
	Subject    string
}

// ToMail renders an email verification message.
func (n VerifyEmail) ToMail(ctx context.Context, notifiable any) (MailMessage, error) {
	user, ok := notifiable.(cauth.MustVerifyEmail)

	if !ok {
		return MailMessage{}, errors.New("notifications: notifiable cannot verify email")
	}

	actionURL := ""

	if n.URLBuilder != nil {
		url, err := n.URLBuilder(ctx, user)

		if err != nil {
			return MailMessage{}, err
		}

		actionURL = url
	}

	subject := n.Subject

	if subject == "" {
		subject = "Verify Email Address"
	}

	return MailMessage{
		Subject:    subject,
		IntroLines: []string{"Please click the button below to verify your email address."},
		ActionText: "Verify Email Address",
		ActionURL:  actionURL,
		OutroLines: []string{"If you did not create an account, no further action is required."},
	}, nil
}

func (n VerifyEmail) ShouldQueue() bool { return false }
func (n VerifyEmail) QueueConnection() string {
	return ""
}
func (n VerifyEmail) QueueName() string {
	return ""
}

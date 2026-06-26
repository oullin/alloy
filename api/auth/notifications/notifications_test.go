package notifications_test

import (
	"context"
	"testing"
	"time"

	"github.com/oullin/alloy/api/auth/notifications"
	cauth "github.com/oullin/alloy/api/contracts/auth"
)

type recordingMailer struct {
	messages []notifications.MailMessage
}

type recordingQueuer struct {
	jobs       []any
	connection string
	queue      string
}

type notifiable struct {
	email string
}

type resettableNotifiable struct {
	notifiable
}

type verifiableNotifiable struct {
	notifiable
	verified bool
}

type queueableNotification struct {
	queue bool
}

type directNotification struct{}

func TestSenderSendsMailToNotifiableRoute(t *testing.T) {
	mailer := &recordingMailer{}
	sender := notifications.NewSender(mailer)
	user := &resettableNotifiable{notifiable: notifiable{email: "taylor@example.com"}}
	notification := notifications.ResetPassword{
		Token: "token",
		URLBuilder: func(_ context.Context, user cauth.CanResetPassword, token string) (string, error) {
			return "https://example.com/reset?email=" + user.GetEmailForPasswordReset() + "&token=" + token, nil
		},
	}

	if err := sender.SendMail(context.Background(), user, notification); err != nil {
		t.Fatal(err)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(mailer.messages))
	}

	message := mailer.messages[0]

	if message.To[0] != "taylor@example.com" {
		t.Fatalf("to = %#v", message.To)
	}

	if message.ActionURL != "https://example.com/reset?email=taylor@example.com&token=token" {
		t.Fatalf("action url = %q", message.ActionURL)
	}

	if message.Subject != "Reset Password Notification" {
		t.Fatalf("subject = %q", message.Subject)
	}
}

func TestSenderQueuesQueueableNotification(t *testing.T) {
	queuer := &recordingQueuer{}
	sender := notifications.NewSender(&recordingMailer{}).WithQueuer(queuer)
	notification := queueableNotification{}

	if err := sender.SendMail(context.Background(), notifiable{email: "taylor@example.com"}, notification); err != nil {
		t.Fatal(err)
	}

	if len(queuer.jobs) != 1 {
		t.Fatalf("jobs = %d, want 1", len(queuer.jobs))
	}

	if queuer.connection != "redis" || queuer.queue != "mail" {
		t.Fatalf("queue route = %s/%s", queuer.connection, queuer.queue)
	}
}

func TestSenderRequiresMailRoute(t *testing.T) {
	err := notifications.NewSender(&recordingMailer{}).SendMail(context.Background(), struct{}{}, directNotification{})

	if err != notifications.ErrMissingMailRoute {
		t.Fatalf("err = %v, want ErrMissingMailRoute", err)
	}
}

func TestVerifyEmailBuildsMailMessage(t *testing.T) {
	user := &verifiableNotifiable{notifiable: notifiable{email: "taylor@example.com"}}
	notification := notifications.VerifyEmail{
		URLBuilder: func(_ context.Context, user cauth.MustVerifyEmail) (string, error) {
			return "https://example.com/verify?email=" + user.GetEmailForVerification(), nil
		},
	}

	message, err := notification.ToMail(context.Background(), user)

	if err != nil {
		t.Fatal(err)
	}

	if message.ActionText != "Verify Email Address" {
		t.Fatalf("action text = %q", message.ActionText)
	}

	if message.ActionURL != "https://example.com/verify?email=taylor@example.com" {
		t.Fatalf("action url = %q", message.ActionURL)
	}
}

func (m *recordingMailer) Send(_ context.Context, message notifications.MailMessage) error {
	m.messages = append(m.messages, message)

	return nil
}

func (q *recordingQueuer) Push(_ context.Context, job any, connection, queue string) error {
	q.jobs = append(q.jobs, job)
	q.connection = connection
	q.queue = queue

	return nil
}

func (n notifiable) RouteNotificationForMail() string { return n.email }

func (n resettableNotifiable) GetEmailForPasswordReset() string { return n.email }

func (n *verifiableNotifiable) HasVerifiedEmail() bool          { return n.verified }
func (n *verifiableNotifiable) MarkEmailAsVerified(time.Time)   { n.verified = true }
func (n *verifiableNotifiable) MarkEmailAsUnverified()          { n.verified = false }
func (n *verifiableNotifiable) GetEmailForVerification() string { return n.email }

func (n queueableNotification) ToMail(context.Context, any) (notifications.MailMessage, error) {
	return notifications.MailMessage{Subject: "Queued"}, nil
}

func (n queueableNotification) ShouldQueue() bool {
	return true
}
func (n queueableNotification) QueueConnection() string { return "redis" }
func (n queueableNotification) QueueName() string       { return "mail" }

func (n directNotification) ToMail(context.Context, any) (notifications.MailMessage, error) {
	return notifications.MailMessage{Subject: "Direct"}, nil
}

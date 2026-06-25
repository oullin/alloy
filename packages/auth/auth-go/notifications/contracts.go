package notifications

import "context"

// MailMessage is the serializable mail payload produced by notifications.
type MailMessage struct {
	To         []string
	Subject    string
	Greeting   string
	IntroLines []string
	ActionText string
	ActionURL  string
	OutroLines []string
	Text       string
	HTML       string
	Headers    map[string]string
}

// Mailer sends a mail message.
type Mailer interface {
	Send(ctx context.Context, message MailMessage) error
}

// Notification can render itself for mail delivery.
type Notification interface {
	ToMail(ctx context.Context, notifiable any) (MailMessage, error)
}

// Notifiable routes notifications to delivery channels.
type Notifiable interface {
	RouteNotificationForMail() string
}

// QueueableNotification exposes queue routing metadata without coupling auth to
// a concrete queue implementation.
type QueueableNotification interface {
	ShouldQueue() bool
	QueueConnection() string
	QueueName() string
}

// Queuer accepts notification delivery jobs.
type Queuer interface {
	Push(ctx context.Context, job any, connection, queue string) error
}

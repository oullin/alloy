package notifications

import (
	"context"
	"errors"
)

// Sender delivers notifications through the configured channels.
type Sender struct {
	mailer Mailer
	queuer Queuer
}

// NewSender creates a Sender.

// WithQueuer configures queue-compatible delivery.

// SendMail sends a notification's mail representation.

// DeliveryJob is a queue-compatible notification delivery payload.
type DeliveryJob struct {
	Notifiable   any
	Notification Notification
}

var (
	ErrMissingMailRoute = errors.New("notifications: missing mail route")
	ErrMissingMailer    = errors.New("notifications: missing mailer")
)

func NewSender(mailer Mailer) *Sender {
	return &Sender{mailer: mailer}
}

func (s *Sender) WithQueuer(queuer Queuer) *Sender {
	s.queuer = queuer

	return s
}

func (s *Sender) SendMail(ctx context.Context, notifiable any, notification Notification) error {
	if queued, ok := notification.(QueueableNotification); ok && queued.ShouldQueue() && s.queuer != nil {
		return s.queuer.Push(ctx, DeliveryJob{
			Notifiable:   notifiable,
			Notification: notification,
		}, queued.QueueConnection(), queued.QueueName())
	}

	if s.mailer == nil {
		return ErrMissingMailer
	}

	message, err := notification.ToMail(ctx, notifiable)

	if err != nil {
		return err
	}

	route := routeForMail(notifiable)

	if route == "" && len(message.To) == 0 {
		return ErrMissingMailRoute
	}

	if route != "" && len(message.To) == 0 {
		message.To = []string{route}
	}

	return s.mailer.Send(ctx, message)
}

func routeForMail(notifiable any) string {
	if routed, ok := notifiable.(Notifiable); ok {
		return routed.RouteNotificationForMail()
	}

	return ""
}

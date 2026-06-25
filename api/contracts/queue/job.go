package queue

import (
	"context"
	"time"
)

// Job represents a queued job instance.
type Job interface {
	UUID() string
	GetJobID() string
	Payload() []byte
	Fire(ctx context.Context) error
	Release(delay time.Duration) error
	Delete() error
	Fail(err error) error
	MarkAsFailed(err error) error
	Attempts() int
	MaxTries() int
	MaxExceptions() int
	Timeout() time.Duration
	Backoff() []time.Duration
	RetryUntil() *time.Time
	IsDeleted() bool
	IsReleased() bool
	HasFailed() bool
	GetQueue() string
	GetConnectionName() string
}

// Handler handles a job type.
type Handler interface {
	Handle(ctx context.Context, job Job) error
}

// HandlerFunc is a function that implements Handler.
type HandlerFunc func(ctx context.Context, job Job) error

// Handle implements Handler.

// FailureHandler is the optional contract a Handler can implement to receive
// a callback when a job failed.
type FailureHandler interface {
	Failed(ctx context.Context, job Job, err error)
}

// JobOptions configures job dispatch options.
type JobOptions struct {
	Backend                 string
	Connection              string
	Delay                   time.Duration
	BatchID                 string
	MaxTries                int
	MaxExceptions           int
	Timeout                 time.Duration
	Backoff                 []time.Duration
	RetryUntil              time.Time
	FailOnTimeout           bool
	UniqueFor               time.Duration
	DeleteWhenMissingModels bool
}

func (f HandlerFunc) Handle(ctx context.Context, job Job) error { return f(ctx, job) }

// WithoutDelay returns a copy of opts with its dispatch delay cleared.
func (o JobOptions) WithoutDelay() JobOptions {
	o.Delay = 0

	return o
}

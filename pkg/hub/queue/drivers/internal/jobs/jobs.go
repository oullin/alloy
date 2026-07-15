// Package jobs holds the state every queue driver's job shares.
//
// It lives under internal/ so each driver subpackage can embed Base without
// widening the public API: Go's internal rule lets anything rooted at
// queue/drivers/ import it, and nothing outside can.
//
// Base's fields stay unexported. Drivers set the immutable ones through
// Config at construction and attach behaviour through the On* setters, so a
// driver can never reach in and desynchronise the released/deleted/failed
// flags from the callbacks that back them.
package jobs

import (
	"context"
	"time"
)

// Config carries the values a driver knows when it pops a job. The zero value
// is meaningful: drivers populate only the fields their backend supplies.
type Config struct {
	ID            string
	UUID          string
	Payload       []byte
	Queue         string
	Connection    string
	Attempts      int
	MaxTries      int
	MaxExceptions int
	Timeout       time.Duration
	Backoff       []time.Duration
	RetryUntil    *time.Time
}

// Base provides default implementations of the queue.Job methods for embedding
// in a driver's own job type.
type Base struct {
	id            string
	uuid          string
	payload       []byte
	queue         string
	connection    string
	attempts      int
	maxTries      int
	maxExceptions int
	timeout       time.Duration
	backoff       []time.Duration
	retryUntil    *time.Time
	released      bool
	deleted       bool
	failed        bool
	releaseFunc   func(delay time.Duration) error
	deleteFunc    func() error
	failFunc      func(err error) error
	fireFunc      func(ctx context.Context) error
}

// New builds a Base from cfg. Attach behaviour with the On* setters.
func New(cfg Config) Base {
	return Base{
		id:            cfg.ID,
		uuid:          cfg.UUID,
		payload:       cfg.Payload,
		queue:         cfg.Queue,
		connection:    cfg.Connection,
		attempts:      cfg.Attempts,
		maxTries:      cfg.MaxTries,
		maxExceptions: cfg.MaxExceptions,
		timeout:       cfg.Timeout,
		backoff:       cfg.Backoff,
		retryUntil:    cfg.RetryUntil,
	}
}

// OnRelease sets the handler backing Release. A job with no handler still
// records that it was released.
func (j *Base) OnRelease(fn func(delay time.Duration) error) { j.releaseFunc = fn }

// OnDelete sets the handler backing Delete.
func (j *Base) OnDelete(fn func() error) { j.deleteFunc = fn }

// OnFail sets the handler backing Fail.
func (j *Base) OnFail(fn func(err error) error) { j.failFunc = fn }

// OnFire sets the handler backing Fire.
func (j *Base) OnFire(fn func(ctx context.Context) error) { j.fireFunc = fn }

func (j *Base) UUID() string              { return j.uuid }
func (j *Base) GetJobID() string          { return j.id }
func (j *Base) Payload() []byte           { return j.payload }
func (j *Base) Attempts() int             { return j.attempts }
func (j *Base) MaxTries() int             { return j.maxTries }
func (j *Base) MaxExceptions() int        { return j.maxExceptions }
func (j *Base) Timeout() time.Duration    { return j.timeout }
func (j *Base) Backoff() []time.Duration  { return j.backoff }
func (j *Base) RetryUntil() *time.Time    { return j.retryUntil }
func (j *Base) IsDeleted() bool           { return j.deleted }
func (j *Base) IsReleased() bool          { return j.released }
func (j *Base) HasFailed() bool           { return j.failed }
func (j *Base) GetQueue() string          { return j.queue }
func (j *Base) GetConnectionName() string { return j.connection }

func (j *Base) Fire(ctx context.Context) error {
	if j.fireFunc != nil {
		return j.fireFunc(ctx)
	}

	return nil
}

func (j *Base) Release(delay time.Duration) error {
	j.released = true

	if j.releaseFunc != nil {
		return j.releaseFunc(delay)
	}

	return nil
}

func (j *Base) Delete() error {
	j.deleted = true

	if j.deleteFunc != nil {
		return j.deleteFunc()
	}

	return nil
}

func (j *Base) Fail(err error) error {
	j.failed = true

	if j.failFunc != nil {
		return j.failFunc(err)
	}

	return nil
}

func (j *Base) MarkAsFailed(err error) error {
	return j.Fail(err)
}

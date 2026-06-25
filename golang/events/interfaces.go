package events

import (
	"context"
	"time"
)

// Listener handles an event. The returned value is collected by Dispatch and
// used by Until to halt on the first non-nil response.
type Listener func(ctx context.Context, event any) (any, error)

// Subscriber registers one or more event-listener mappings on a dispatcher.
type Subscriber interface {
	Subscribe(dispatcher Dispatcher)
}

// Dispatcher dispatches domain events and manages listeners.
type Dispatcher interface {
	Listen(events any, listeners ...Listener)
	HasListeners(event any) bool
	HasWildcardListeners(event any) bool
	Subscribe(subscriber Subscriber)
	Until(ctx context.Context, event any) (any, error)
	Dispatch(ctx context.Context, event any) ([]any, error)
	Push(ctx context.Context, event any)
	Flush(ctx context.Context, event string) error
	Forget(event any)
	ForgetPushed()
	GetListeners(event any) []Listener
}

// ShouldQueue is a marker interface. Events or listener wrappers implementing
// this are dispatched to the queue backend instead of executing inline.
type ShouldQueue interface {
	ShouldQueue()
}

// ShouldDispatchAfterCommit is a marker interface for events that should be
// deferred until the active database transaction commits.
type ShouldDispatchAfterCommit interface {
	ShouldDispatchAfterCommit()
}

// ShouldHandleEventsAfterCommit is a marker interface for listeners that
// should defer execution until the active database transaction commits.
type ShouldHandleEventsAfterCommit interface {
	ShouldHandleEventsAfterCommit()
}

// TransactionManager allows the dispatcher to defer events until a database
// transaction commits.
type TransactionManager interface {
	AfterCommit(fn func())
}

// QueueResolver creates a queue-like backend on demand for dispatching
// queued listeners.
type QueueResolver func() QueueBackend

// TransactionManagerResolver creates a TransactionManager on demand.
type TransactionManagerResolver func() TransactionManager

// QueueBackend is the minimal interface required to push listener jobs.
type QueueBackend interface {
	Push(queue string, payload []byte) error
	PushDelayed(queue string, payload []byte, delay time.Duration) error
}

// ListenerOptions configures a queued listener.
type ListenerOptions struct {
	Connection    string
	Queue         string
	Delay         time.Duration
	Tries         int
	MaxExceptions int
	Timeout       time.Duration
	Backoff       []time.Duration
	AfterCommit   *bool
}

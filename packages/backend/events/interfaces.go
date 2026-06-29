package events

import cevents "alloy.dev/backend/contracts/events"

// Listener handles an event. The returned value is collected by Dispatch and
// used by Until to halt on the first non-nil response.
type Listener = cevents.Listener

// Subscriber registers one or more event-listener mappings on a dispatcher.
type Subscriber = cevents.Subscriber

// Dispatcher dispatches domain events and manages listeners.
type Dispatcher = cevents.Dispatcher

// ShouldQueue is a marker interface. Events or listener wrappers implementing
// this are dispatched to the queue backend instead of executing inline.
type ShouldQueue = cevents.ShouldQueue

// ShouldDispatchAfterCommit is a marker interface for events that should be
// deferred until the active database transaction commits.
type ShouldDispatchAfterCommit = cevents.ShouldDispatchAfterCommit

// ShouldHandleEventsAfterCommit is a marker interface for listeners that
// should defer execution until the active database transaction commits.
type ShouldHandleEventsAfterCommit = cevents.ShouldHandleEventsAfterCommit

// TransactionManager allows the dispatcher to defer events until a database
// transaction commits.
type TransactionManager = cevents.TransactionManager

// QueueResolver creates a queue-like backend on demand for dispatching
// queued listeners.
type QueueResolver = cevents.QueueResolver

// TransactionManagerResolver creates a TransactionManager on demand.
type TransactionManagerResolver = cevents.TransactionManagerResolver

// QueueBackend is the minimal interface required to push listener jobs.
type QueueBackend = cevents.QueueBackend

// ListenerOptions configures a queued listener.
type ListenerOptions = cevents.ListenerOptions

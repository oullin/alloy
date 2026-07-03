package bus

import "errors"

var (
	// ErrNoQueueBackend is returned when a queued dispatch is attempted and
	// no queue backend has been configured on the dispatcher.
	ErrNoQueueBackend = errors.New("bus: no queue backend configured")

	// ErrNoBatchRepository is returned when a batch operation is attempted
	// and no batch repository has been configured on the dispatcher.
	ErrNoBatchRepository = errors.New("bus: no batch repository configured")

	// ErrNoHandler is returned when a command has no registered handler and
	// does not implement its own Handle method.
	ErrNoHandler = errors.New("bus: no handler registered")

	// ErrEmptyChain is returned when a chain with no jobs is dispatched.
	ErrEmptyChain = errors.New("bus: cannot dispatch an empty chain")

	// ErrInvalidChainHead is returned when the first job of a chain cannot
	// carry the remaining jobs because it has no Queueable Chain support.
	ErrInvalidChainHead = errors.New("bus: first chained job cannot accept remaining jobs")

	// ErrBatchNotFound is returned when a batch ID does not exist in the
	// repository.
	ErrBatchNotFound = errors.New("bus: batch not found")
)

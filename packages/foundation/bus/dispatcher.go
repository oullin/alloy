package bus

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/oullin/alloy/packages/foundation/bus/pipeline"
	"github.com/oullin/alloy/packages/foundation/queue"
)

// DispatcherImpl is the concrete implementation of QueueingDispatcher.
type DispatcherImpl struct {
	mu                     sync.RWMutex
	handlers               map[reflect.Type]Handler
	pipes                  []Pipe
	deferred               []any
	queueBackend           queue.Backend
	batchRepo              BatchRepository
	eventFunc              EventFunc
	dispatchAfterResponses bool
}

// NewDispatcher creates a DispatcherImpl.
// queueBackend may be nil if DispatchToQueue is never called.
func NewDispatcher(queueBackend queue.Backend, batchRepo BatchRepository) *DispatcherImpl {
	return &DispatcherImpl{
		handlers:     make(map[reflect.Type]Handler),
		queueBackend: queueBackend,
		batchRepo:    batchRepo,
	}
}

// Map registers a command type → handler mapping.
// command should be a zero-value instance of the command struct (e.g. MyCommand{}).
func (d *DispatcherImpl) Map(command any, handler Handler) Dispatcher {
	d.mu.Lock()

	defer d.mu.Unlock()

	t := reflect.TypeOf(command)
	d.handlers[t] = handler

	return d
}

// PipeThrough replaces the current pipe list with the given pipes.
func (d *DispatcherImpl) PipeThrough(pipes ...Pipe) Dispatcher {
	d.mu.Lock()

	defer d.mu.Unlock()

	d.pipes = pipes

	return d
}

// CommandShouldBeQueued reports whether the command implements ShouldQueue.
func (d *DispatcherImpl) CommandShouldBeQueued(command any) bool {
	_, ok := command.(ShouldQueue)

	return ok
}

// WithDispatchingAfterResponses enables after-response dispatch mode.
func (d *DispatcherImpl) WithDispatchingAfterResponses() *DispatcherImpl {
	d.mu.Lock()

	defer d.mu.Unlock()

	d.dispatchAfterResponses = true

	return d
}

// WithoutDispatchingAfterResponses disables after-response dispatch mode.
func (d *DispatcherImpl) WithoutDispatchingAfterResponses() *DispatcherImpl {
	d.mu.Lock()

	defer d.mu.Unlock()

	d.dispatchAfterResponses = false

	return d
}

// SetEventFunc sets the event callback for batch lifecycle events.
func (d *DispatcherImpl) SetEventFunc(fn EventFunc) {
	d.eventFunc = fn
}

// Dispatch sends the command through the pipeline and executes it.
// If the command implements ShouldQueue, it is dispatched to the queue.
// If dispatchAfterResponses is enabled, the command is deferred.
func (d *DispatcherImpl) Dispatch(ctx context.Context, command any) (any, error) {
	d.mu.RLock()
	afterResponses := d.dispatchAfterResponses
	d.mu.RUnlock()

	if afterResponses {
		return nil, d.DispatchAfterResponse(ctx, command)
	}

	if d.CommandShouldBeQueued(command) {
		return nil, d.DispatchToQueue(ctx, command)
	}

	return d.runThroughPipeline(ctx, command)
}

// DispatchSync executes the command synchronously, bypassing any queue routing.
func (d *DispatcherImpl) DispatchSync(ctx context.Context, command any) (any, error) {
	return d.execute(ctx, command)
}

// DispatchNow is an alias for DispatchSync.
func (d *DispatcherImpl) DispatchNow(ctx context.Context, command any) (any, error) {
	return d.DispatchSync(ctx, command)
}

// DispatchAfterResponse buffers a command for execution after the response is sent.
// Call FlushDeferred() to process all buffered commands.
func (d *DispatcherImpl) DispatchAfterResponse(_ context.Context, command any) error {
	d.mu.Lock()

	defer d.mu.Unlock()

	d.deferred = append(d.deferred, command)

	return nil
}

// FlushDeferred dispatches all after-response commands synchronously.
func (d *DispatcherImpl) FlushDeferred(ctx context.Context) error {
	d.mu.Lock()
	commands := d.deferred
	d.deferred = nil
	d.mu.Unlock()

	for _, cmd := range commands {
		if _, err := d.execute(ctx, cmd); err != nil {
			return err
		}
	}

	return nil
}

// DispatchToQueue serialises the command and pushes it to the queue backend.
func (d *DispatcherImpl) DispatchToQueue(ctx context.Context, command any) error {
	if d.queueBackend == nil {
		return ErrNoQueueBackend
	}

	payload, err := marshalCommand(command)

	if err != nil {
		return err
	}

	queueName := "default"

	if q, ok := command.(interface{ GetQueue() string }); ok {
		queueName = q.GetQueue()
	}

	if delayer, ok := command.(interface{ GetDelay() time.Duration }); ok {
		if delay := delayer.GetDelay(); delay > 0 {
			_, err = d.queueBackend.PushDelayed(ctx, queueName, payload, delay)

			return err
		}
	}

	_, err = d.queueBackend.Push(ctx, queueName, payload)

	return err
}

// FindBatch retrieves a Batch by ID from the repository.
func (d *DispatcherImpl) FindBatch(ctx context.Context, id string) (*Batch, error) {
	if d.batchRepo == nil {
		return nil, ErrNoBatchRepository
	}

	return d.batchRepo.Get(ctx, id)
}

// HasCommandHandler reports whether a handler is registered for the command type.
func (d *DispatcherImpl) HasCommandHandler(command any) bool {
	d.mu.RLock()

	defer d.mu.RUnlock()

	_, ok := d.handlers[reflect.TypeOf(command)]

	return ok
}

// GetCommandHandler returns the handler for the given command type.
func (d *DispatcherImpl) GetCommandHandler(command any) (Handler, bool) {
	d.mu.RLock()

	defer d.mu.RUnlock()

	h, ok := d.handlers[reflect.TypeOf(command)]

	return h, ok
}

// Chain creates a PendingChain for sequential job execution.
func (d *DispatcherImpl) Chain(jobs []any) *PendingChain {
	return NewPendingChain(d, jobs)
}

// Batch creates a PendingBatch for the given jobs.
func (d *DispatcherImpl) Batch(jobs []any) *PendingBatch {
	pb := NewPendingBatch(d, jobs)
	pb.batchRepo = d.batchRepo
	pb.eventFunc = d.eventFunc

	return pb
}

func (d *DispatcherImpl) runThroughPipeline(ctx context.Context, command any) (any, error) {
	d.mu.RLock()
	pipes := d.pipes
	d.mu.RUnlock()

	// Convert bus.Pipe to pipeline.Pipe.
	pipelinePipes := make([]pipeline.Pipe, len(pipes))

	for i, p := range pipes {
		p := p // capture loop variable.
		pipelinePipes[i] = func(ctx context.Context, cmd any, next pipeline.Handler) (any, error) {
			return p(ctx, cmd, func(ctx context.Context, cmd any) (any, error) {
				return next(ctx, cmd)
			})
		}
	}

	return pipeline.New().
		Through(pipelinePipes...).
		Send(ctx, command, func(ctx context.Context, cmd any) (any, error) {
			return d.execute(ctx, cmd)
		})
}

func (d *DispatcherImpl) execute(ctx context.Context, command any) (any, error) {
	d.mu.RLock()
	handler, ok := d.handlers[reflect.TypeOf(command)]
	d.mu.RUnlock()

	if !ok {
		// If the command implements Handler itself, call it.
		if h, ok := command.(interface {
			Handle(ctx context.Context) (any, error)
		}); ok {
			return h.Handle(ctx)
		}

		return nil, fmt.Errorf("%w for %T", ErrNoHandler, command)
	}

	return handler(ctx, command)
}

func marshalCommand(command any) ([]byte, error) {
	p, err := marshalJSON(command)

	if err != nil {
		return nil, fmt.Errorf("bus: marshal command: %w", err)
	}

	return p, nil
}

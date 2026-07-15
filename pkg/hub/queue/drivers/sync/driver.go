package sync

import (
	"context"
	"time"

	"github.com/oullin/alloy/pkg/hub/queue"
	"github.com/oullin/alloy/pkg/hub/queue/drivers/internal/jobs"
)

// Driver executes jobs immediately in the same goroutine. It is
// The Go API differs from the upstream in one ergonomic way: the handler
// that runs every job is injected at construction time rather than
// resolved from a container per-push. This keeps the existing bedrock
// Backend interface (Push takes raw []byte) usable as-is and matches
// the upstream semantics when a single handler is effectively registered
// for the connection.
//
// Events: if an EventEmitter is wired up via SetEmitter, every Push
// fires the upstream event sequence:
//
//	JobProcessing → handler.Handle → JobProcessed → JobAttempted
//	JobProcessing → handler.Handle → JobExceptionOccurred → Fail → JobFailed → JobAttempted
//
// If the handler also implements queue.FailureHandler, its Failed
// method is invoked between the Fail call and the JobFailed emission,
type Driver struct {
	connection string
	handler    queue.Handler
	emitter    queue.EventEmitter
	tx         TransactionCallbackRegistrar
}

type syncJob struct{ jobs.Base }

// TransactionCallbackRegistrar is the small transaction hook Driver
// needs to mirror the upstream after-commit dispatch registration.
type TransactionCallbackRegistrar interface {
	AddCallback(func())
}

// NewDriver creates a Driver. handler is called synchronously
// for every Push. The emitter is nil until SetEmitter is called.
func NewDriver(connection string, handler queue.Handler) *Driver {
	return &Driver{connection: connection, handler: handler}
}

// SetEmitter installs an EventEmitter. A nil emitter disables event
// emission (the default). Returns the driver for chaining.
func (d *Driver) SetEmitter(e queue.EventEmitter) *Driver {
	d.emitter = e

	return d
}

func (d *Driver) SetTransactionManager(tx TransactionCallbackRegistrar) *Driver {
	d.tx = tx

	return d
}

func (d *Driver) Push(ctx context.Context, queueName string, payload []byte) (string, error) {
	return d.PushJob(ctx, queueName, nil, payload, nil)
}

// PushJob dispatches payload with the original job value and connection
// config available for after-commit precedence checks.
func (d *Driver) PushJob(ctx context.Context, queueName string, jobValue any, payload []byte, config map[string]any) (string, error) {
	job := &syncJob{Base: jobs.New(jobs.Config{Payload: payload, Queue: queueName, Connection: d.connection})}
	job.OnFire(func(ctx context.Context) error {
		return d.handler.Handle(ctx, job)
	})

	if d.tx != nil && queue.ShouldDispatchAfterCommit(jobValue, config) {
		d.tx.AddCallback(func() {
			_ = d.executeJob(ctx, job)
		})

		return "", nil
	}

	return "", d.executeJob(ctx, job)
}

// executeJob runs the upstream-style lifecycle around the
// handler.Handle call and returns whatever error the handler produced.
func (d *Driver) executeJob(ctx context.Context, job queue.Job) error {
	d.emit(queue.JobProcessing{ConnectionName: d.connection, Job: job})

	err := job.Fire(ctx)

	if err != nil {
		d.emit(queue.JobExceptionOccurred{ConnectionName: d.connection, Job: job, Err: err})

		_ = job.Fail(err)

		if fh, ok := d.handler.(queue.FailureHandler); ok {
			fh.Failed(ctx, job, err)
		}

		d.emit(queue.JobFailed{ConnectionName: d.connection, Job: job, Err: err})
		d.emit(queue.JobAttempted{ConnectionName: d.connection, Job: job})

		return err
	}

	d.emit(queue.JobProcessed{ConnectionName: d.connection, Job: job})
	d.emit(queue.JobAttempted{ConnectionName: d.connection, Job: job})

	return nil
}

func (d *Driver) emit(event any) {
	if d.emitter != nil {
		d.emitter.Emit(event)
	}
}

func (d *Driver) PushDelayed(ctx context.Context, queueName string, payload []byte, delay time.Duration) (string, error) {
	time.Sleep(delay)

	return d.Push(ctx, queueName, payload)
}

func (d *Driver) PushMultiple(ctx context.Context, queueName string, payloads [][]byte) ([]string, error) {
	ids := make([]string, 0, len(payloads))

	for _, p := range payloads {
		id, err := d.Push(ctx, queueName, p)

		if err != nil {
			return ids, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (d *Driver) Pop(_ context.Context, _ string) (queue.Job, error) { return nil, queue.ErrNoJob }

func (d *Driver) Size(_ context.Context, _ string) (int64, error)         { return 0, nil }
func (d *Driver) PendingSize(_ context.Context, _ string) (int64, error)  { return 0, nil }
func (d *Driver) DelayedSize(_ context.Context, _ string) (int64, error)  { return 0, nil }
func (d *Driver) ReservedSize(_ context.Context, _ string) (int64, error) { return 0, nil }
func (d *Driver) ConnectionName() string                                  { return d.connection }

// QueueNames returns an empty slice — sync jobs are executed inline
// and never persisted, so there is nothing to enumerate.
func (d *Driver) QueueNames(_ context.Context) ([]string, error) { return nil, nil }

// PendingJobs returns an empty slice. The sync driver never has
// pending work — jobs are processed inside the Push call itself.
func (d *Driver) PendingJobs(_ context.Context, _ string) ([]queue.InspectedJob, error) {
	return nil, nil
}

// DelayedJobs returns an empty slice.
func (d *Driver) DelayedJobs(_ context.Context, _ string) ([]queue.InspectedJob, error) {
	return nil, nil
}

// ReservedJobs returns an empty slice.
func (d *Driver) ReservedJobs(_ context.Context, _ string) ([]queue.InspectedJob, error) {
	return nil, nil
}

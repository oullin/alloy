package null

import (
	"context"
	"time"

	"github.com/oullin/alloy/pkg/hub/queue"
)

// Driver discards all jobs silently.
type Driver struct {
	connection string
}

// NewDriver creates a Driver.
func NewDriver(connection string) *Driver {
	return &Driver{connection: connection}
}

func (d *Driver) Push(_ context.Context, _ string, _ []byte) (string, error) { return "", nil }
func (d *Driver) PushDelayed(_ context.Context, _ string, _ []byte, _ time.Duration) (string, error) {
	return "", nil
}
func (d *Driver) PushMultiple(_ context.Context, _ string, payloads [][]byte) ([]string, error) {
	return make([]string, len(payloads)), nil
}
func (d *Driver) Pop(_ context.Context, _ string) (queue.Job, error)      { return nil, queue.ErrNoJob }
func (d *Driver) Size(_ context.Context, _ string) (int64, error)         { return 0, nil }
func (d *Driver) PendingSize(_ context.Context, _ string) (int64, error)  { return 0, nil }
func (d *Driver) DelayedSize(_ context.Context, _ string) (int64, error)  { return 0, nil }
func (d *Driver) ReservedSize(_ context.Context, _ string) (int64, error) { return 0, nil }
func (d *Driver) ConnectionName() string                                  { return d.connection }

// QueueNames returns an empty slice — the null driver never holds jobs
// and therefore has no queues to enumerate.
func (d *Driver) QueueNames(_ context.Context) ([]string, error) { return nil, nil }

// PendingJobs returns an empty slice.
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

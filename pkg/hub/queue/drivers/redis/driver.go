package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/oullin/alloy/pkg/hub/queue"
	"github.com/oullin/alloy/pkg/hub/queue/drivers/internal/jobs"
)

// Driver stores jobs in Redis lists and sorted sets.
type Driver struct {
	client     Client
	connection string
}

type job struct{ jobs.Base }

func NewDriver(client Client, connection string) *Driver {
	return &Driver{client: client, connection: connection}
}

func (d *Driver) Push(ctx context.Context, queueName string, payload []byte) (string, error) {
	return "", d.client.LPush(ctx, queueKey(queueName), string(payload))
}

func (d *Driver) PushDelayed(ctx context.Context, queueName string, payload []byte, delay time.Duration) (string, error) {
	score := float64(time.Now().Add(delay).Unix())

	return "", d.client.ZAdd(ctx, delayedKey(queueName), score, string(payload))
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

func (d *Driver) Pop(ctx context.Context, queueName string) (queue.Job, error) {

	d.migrateDue(ctx, queueName)

	raw, err := d.client.RPop(ctx, queueKey(queueName))

	if err != nil || raw == "" {
		return nil, queue.ErrNoJob
	}

	job := &job{
		Base: jobs.New(jobs.Config{
			Payload:    []byte(raw),
			Queue:      queueName,
			Connection: d.connection,
		}),
	}

	job.OnDelete(func() error { return nil })
	job.OnRelease(func(delay time.Duration) error {
		if delay > 0 {
			_, err := d.PushDelayed(ctx, queueName, []byte(raw), delay)

			return err
		}

		_, err := d.Push(ctx, queueName, []byte(raw))

		return err
	})

	job.OnFail(func(err error) error {
		errMsg := ""

		if err != nil {
			errMsg = err.Error()
		}

		failed := map[string]string{"exception": errMsg, "payload": raw}
		b, _ := json.Marshal(failed)

		return d.client.LPush(ctx, failedKey(queueName), string(b))
	})

	return job, nil
}

func (d *Driver) Size(ctx context.Context, queueName string) (int64, error) {
	return d.client.LLen(ctx, queueKey(queueName))
}

func (d *Driver) ClearQueue(ctx context.Context, queueName string) error {
	deleter, ok := d.client.(Deleter)

	if !ok {
		return nil
	}

	_, err := deleter.Del(ctx, queueKey(queueName), delayedKey(queueName), failedKey(queueName))

	return err
}

func (d *Driver) PendingSize(ctx context.Context, queueName string) (int64, error) {
	return d.client.LLen(ctx, queueKey(queueName))
}

func (d *Driver) DelayedSize(ctx context.Context, queueName string) (int64, error) {
	return d.client.ZCard(ctx, delayedKey(queueName))
}

func (d *Driver) ReservedSize(_ context.Context, _ string) (int64, error) {

	return 0, nil
}

// ConnectionName reports the queue connection this driver was built for.
func (d *Driver) ConnectionName() string { return d.connection }

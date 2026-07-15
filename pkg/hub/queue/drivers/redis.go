package drivers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/oullin/alloy/pkg/hub/queue"
)

// RedisDriver stores jobs in Redis lists and sorted sets.
type RedisDriver struct {
	client     RedisClient
	connection string
}

type redisJob struct{ BaseJob }

func NewRedisDriver(client RedisClient, connection string) *RedisDriver {
	return &RedisDriver{client: client, connection: connection}
}

func (d *RedisDriver) Push(ctx context.Context, queueName string, payload []byte) (string, error) {
	return "", d.client.LPush(ctx, queueKey(queueName), string(payload))
}

func (d *RedisDriver) PushDelayed(ctx context.Context, queueName string, payload []byte, delay time.Duration) (string, error) {
	score := float64(time.Now().Add(delay).Unix())

	return "", d.client.ZAdd(ctx, delayedKey(queueName), score, string(payload))
}

func (d *RedisDriver) PushMultiple(ctx context.Context, queueName string, payloads [][]byte) ([]string, error) {
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

func (d *RedisDriver) Pop(ctx context.Context, queueName string) (queue.Job, error) {

	d.migrateDue(ctx, queueName)

	raw, err := d.client.RPop(ctx, queueKey(queueName))

	if err != nil || raw == "" {
		return nil, queue.ErrNoJob
	}

	job := &redisJob{
		BaseJob: BaseJob{
			payload:    []byte(raw),
			queue:      queueName,
			connection: d.connection,
		},
	}
	job.deleteFunc = func() error { return nil }
	job.releaseFunc = func(delay time.Duration) error {
		if delay > 0 {
			_, err := d.PushDelayed(ctx, queueName, []byte(raw), delay)

			return err
		}

		_, err := d.Push(ctx, queueName, []byte(raw))

		return err
	}

	job.failFunc = func(err error) error {
		errMsg := ""

		if err != nil {
			errMsg = err.Error()
		}

		failed := map[string]string{"exception": errMsg, "payload": raw}
		b, _ := json.Marshal(failed)

		return d.client.LPush(ctx, failedKey(queueName), string(b))
	}

	return job, nil
}

func (d *RedisDriver) Size(ctx context.Context, queueName string) (int64, error) {
	return d.client.LLen(ctx, queueKey(queueName))
}

func (d *RedisDriver) ClearQueue(ctx context.Context, queueName string) error {
	deleter, ok := d.client.(RedisDeleter)

	if !ok {
		return nil
	}

	_, err := deleter.Del(ctx, queueKey(queueName), delayedKey(queueName), failedKey(queueName))

	return err
}

func (d *RedisDriver) PendingSize(ctx context.Context, queueName string) (int64, error) {
	return d.client.LLen(ctx, queueKey(queueName))
}

func (d *RedisDriver) DelayedSize(ctx context.Context, queueName string) (int64, error) {
	return d.client.ZCard(ctx, delayedKey(queueName))
}

func (d *RedisDriver) ReservedSize(_ context.Context, _ string) (int64, error) {

	return 0, nil
}

// ConnectionName reports the queue connection this driver was built for.
func (d *RedisDriver) ConnectionName() string { return d.connection }

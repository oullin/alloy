package drivers

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/oullin/alloy/pkg/hub/queue"
)

// RedisDriver stores jobs in Redis lists and sorted sets.
//
// At-Least-Once Semantics:
// The driver provides at-least-once delivery guarantees using a reserved sorted set
// for job visibility timeout. When a worker pops a job, the job is atomically moved to
// the reserved set. If the worker crashes or fails to delete/release the job within
// the visibility timeout, the job is reclaimed and re-queued.
//
// Duplicate Delivery Tradeoff:
// A job can be delivered more than once if:
//  1. A handler takes longer to run than the visibility timeout.
//  2. The worker crashes in the brief window between pushing a job back to the list
//     (release/fail) and removing it from the reserved set.
//
// Therefore, handlers running on this driver should be designed to be idempotent.
// Choosing the right visibility timeout is a tradeoff: a timeout shorter than the
// longest handler runtime leads to double processing, while a timeout too long results
// in slow reclaim of jobs after worker crashes.
type RedisDriver struct {
	client            RedisClient
	connection        string
	visibilityTimeout time.Duration

	// isCluster caches the cluster-connection check result. It is nil
	// the upstream null-coalescing assignment ($this->isCluster ??= ...).
	isCluster *bool
}

type redisJob struct{ BaseJob }

const defaultCleanupTimeout = 5 * time.Second

func cleanupContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultCleanupTimeout)
}

func NewRedisDriver(client RedisClient, connection string) *RedisDriver {
	return &RedisDriver{
		client:            client,
		connection:        connection,
		visibilityTimeout: 60 * time.Second,
	}
}

// SetVisibilityTimeout sets the visibility timeout for reserved jobs.
// Tradeoff: a timeout shorter than the longest handler runtime leads to double processing,
// while a timeout too long results in slow reclaim after worker crashes.
func (d *RedisDriver) SetVisibilityTimeout(timeout time.Duration) {
	d.visibilityTimeout = timeout
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
	d.migrateExpired(ctx, queueName)

	score := float64(time.Now().Add(d.visibilityTimeout).Unix())
	res, err := d.client.Eval(ctx, popAndReserveLua, []string{queueKey(queueName), reservedKey(queueName)}, score)

	var raw string

	var popped bool

	if err == nil {
		switch v := res.(type) {
		case string:
			if v != "" {
				raw = v
				popped = true
			}
		case []byte:
			if len(v) > 0 {
				raw = string(v)
				popped = true
			}
		case nil:
			return nil, queue.ErrNoJob
		}
	}

	var shouldFallback bool

	if err != nil {
		// A cancelled or expired context is terminal: falling back to the
		// non-atomic RPop path would only fail again (or run a second,
		// unnecessary round-trip), so surface the error instead.
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}

		shouldFallback = true
	} else if res != nil {
		switch res.(type) {
		case int, int32, int64:
			shouldFallback = true
		}
	}

	if !popped {
		if shouldFallback {
			// Fallback to non-atomic RPop + ZAdd. Triggered by client implementations that
			// lack Lua script execution support.
			// Exposure: A worker crash between RPop and ZAdd, or after a failed ZAdd,
			// will cause the job to be lost (at-most-once semantics during this window).
			var rpopErr error

			raw, rpopErr = d.client.RPop(ctx, queueKey(queueName))

			// Same reasoning as the Eval path above: a cancelled or expired
			// context is terminal, and reporting it as an empty queue would
			// tell a shutting-down worker it had drained the backlog.
			if errors.Is(rpopErr, context.Canceled) || errors.Is(rpopErr, context.DeadlineExceeded) {
				return nil, rpopErr
			}

			if rpopErr != nil || raw == "" {
				return nil, queue.ErrNoJob
			}

			if zerr := d.client.ZAdd(ctx, reservedKey(queueName), score, raw); zerr != nil {
				// Proceed processing the job, but it is unreserved (at-most-once) if the worker crashes.
			}
		} else {
			return nil, queue.ErrNoJob
		}
	}

	p, pErr := queue.UnmarshalPayload([]byte(raw))

	var attempts int

	if pErr == nil && p != nil {
		attempts = p.Tries
	}

	job := &redisJob{
		BaseJob: BaseJob{
			payload:    []byte(raw),
			queue:      queueName,
			connection: d.connection,
			attempts:   attempts,
		},
	}
	job.deleteFunc = func() error {
		cleanupCtx, cancel := cleanupContext()

		defer cancel()

		return d.client.ZRem(cleanupCtx, reservedKey(queueName), raw)
	}

	job.releaseFunc = func(delay time.Duration) error {
		cleanupCtx, cancel := cleanupContext()

		defer cancel()

		releasedPayload := []byte(raw)

		// Unmarshal into a generic map rather than the strict queue.Payload
		// struct so that any unrecognized or custom top-level fields in the
		// original payload survive the release round-trip; only "tries" is bumped.
		var pMap map[string]any

		if err := json.Unmarshal([]byte(raw), &pMap); err == nil && pMap != nil {
			pMap["tries"] = job.attempts + 1

			if updatedRaw, err := json.Marshal(pMap); err == nil {
				releasedPayload = updatedRaw
			}
		}

		if delay > 0 {
			_, err := d.PushDelayed(cleanupCtx, queueName, releasedPayload, delay)

			if err != nil {
				return err
			}
		} else {
			_, err := d.Push(cleanupCtx, queueName, releasedPayload)

			if err != nil {
				return err
			}
		}

		return d.client.ZRem(cleanupCtx, reservedKey(queueName), raw)
	}

	job.failFunc = func(err error) error {
		cleanupCtx, cancel := cleanupContext()

		defer cancel()

		errMsg := ""

		if err != nil {
			errMsg = err.Error()
		}

		failed := map[string]string{"exception": errMsg, "payload": raw}
		b, _ := json.Marshal(failed)

		if pushErr := d.client.LPush(cleanupCtx, failedKey(queueName), string(b)); pushErr != nil {
			return pushErr
		}

		return d.client.ZRem(cleanupCtx, reservedKey(queueName), raw)
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

	_, err := deleter.Del(ctx, queueKey(queueName), delayedKey(queueName), reservedKey(queueName), failedKey(queueName))

	return err
}

func (d *RedisDriver) PendingSize(ctx context.Context, queueName string) (int64, error) {
	return d.client.LLen(ctx, queueKey(queueName))
}

func (d *RedisDriver) DelayedSize(ctx context.Context, queueName string) (int64, error) {
	return d.client.ZCard(ctx, delayedKey(queueName))
}

func (d *RedisDriver) ReservedSize(ctx context.Context, queueName string) (int64, error) {
	return d.client.ZCard(ctx, reservedKey(queueName))
}

func (d *RedisDriver) ConnectionName() string { return d.connection }

package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oullin/alloy/pkg/hub/queue"
	"github.com/oullin/alloy/pkg/hub/queue/drivers/internal/redistest"
	"github.com/oullin/alloy/pkg/hub/queue/drivers/redis"
)

// deleterRedisClient extends redistest.Client with the optional Deleter
// contract that ClearQueue needs, recording the keys it was asked to delete.
type deleterRedisClient struct {
	*redistest.Client
	deleted []string
	delErr  error
}

func (c *deleterRedisClient) Del(_ context.Context, keys ...string) (int64, error) {
	if c.delErr != nil {
		return 0, c.delErr
	}

	c.deleted = append(c.deleted, keys...)

	return int64(len(keys)), nil
}

func TestRedisDriverPush(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, err := drv.Push(context.Background(), "default", []byte("payload"))

	if err != nil {
		t.Fatal(err)
	}

	n, _ := client.LLen(context.Background(), "queues:default")

	if n != 1 {
		t.Errorf("expected 1 item in list, got %d", n)
	}
}

func TestRedisDriverPushDelayed(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, err := drv.PushDelayed(context.Background(), "default", []byte("payload"), 5*time.Second)

	if err != nil {
		t.Fatal(err)
	}

	n, _ := client.ZCard(context.Background(), "queues:default:delayed")

	if n != 1 {
		t.Errorf("expected 1 delayed, got %d", n)
	}
}

func TestRedisDriverPushMultiple(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	ids, err := drv.PushMultiple(context.Background(), "default", [][]byte{
		[]byte("a"), []byte("b"), []byte("c"),
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != 3 {
		t.Errorf("expected 3 ids, got %d", len(ids))
	}

	n, _ := client.LLen(context.Background(), "queues:default")

	if n != 3 {
		t.Errorf("expected 3 items, got %d", n)
	}
}

func TestRedisDriverPopFromMainQueue(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, _ = drv.Push(context.Background(), "default", []byte("test-payload"))

	job, err := drv.Pop(context.Background(), "default")

	if err != nil {
		t.Fatal(err)
	}

	if string(job.Payload()) != "test-payload" {
		t.Errorf("expected 'test-payload', got %q", job.Payload())
	}

	if job.GetQueue() != "default" {
		t.Errorf("expected queue 'default', got %q", job.GetQueue())
	}

	if job.GetConnectionName() != "redis" {
		t.Errorf("expected connection 'redis', got %q", job.GetConnectionName())
	}
}

func TestRedisDriverPopEmptyQueueReturnsErrNoJob(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, err := drv.Pop(context.Background(), "empty")

	if !errors.Is(err, queue.ErrNoJob) {
		t.Fatalf("expected ErrNoJob, got %v", err)
	}
}

func TestRedisDriverJobRelease(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, _ = drv.Push(context.Background(), "default", []byte("release-me"))

	job, _ := drv.Pop(context.Background(), "default")

	err := job.Release(0)

	if err != nil {
		t.Fatal(err)
	}

	n, _ := client.LLen(context.Background(), "queues:default")

	if n != 1 {
		t.Errorf("expected 1 item after release, got %d", n)
	}
}

func TestRedisDriverJobReleaseWithDelay(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, _ = drv.Push(context.Background(), "default", []byte("delay-me"))

	job, _ := drv.Pop(context.Background(), "default")

	err := job.Release(5 * time.Second)

	if err != nil {
		t.Fatal(err)
	}

	n, _ := client.ZCard(context.Background(), "queues:default:delayed")

	if n != 1 {
		t.Errorf("expected 1 delayed after release, got %d", n)
	}
}

func TestRedisDriverJobDelete(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, _ = drv.Push(context.Background(), "default", []byte("delete-me"))

	job, _ := drv.Pop(context.Background(), "default")

	err := job.Delete()

	if err != nil {
		t.Fatal(err)
	}
}

func TestRedisDriverJobFail(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, _ = drv.Push(context.Background(), "default", []byte("fail-me"))

	job, _ := drv.Pop(context.Background(), "default")

	err := job.Fail(errors.New("something broke"))

	if err != nil {
		t.Fatal(err)
	}

	n, _ := client.LLen(context.Background(), "queues:default:failed")

	if n != 1 {
		t.Errorf("expected 1 failed entry, got %d", n)
	}
}

func TestRedisDriverSize(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, _ = drv.Push(context.Background(), "default", []byte("a"))
	_, _ = drv.Push(context.Background(), "default", []byte("b"))

	n, err := drv.Size(context.Background(), "default")

	if err != nil {
		t.Fatal(err)
	}

	if n != 2 {
		t.Errorf("expected size 2, got %d", n)
	}
}

func TestRedisDriverPendingSize(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, _ = drv.Push(context.Background(), "default", []byte("a"))

	n, err := drv.PendingSize(context.Background(), "default")

	if err != nil {
		t.Fatal(err)
	}

	if n != 1 {
		t.Errorf("expected pending size 1, got %d", n)
	}
}

func TestRedisDriverDelayedSize(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	_, _ = drv.PushDelayed(context.Background(), "default", []byte("a"), time.Hour)

	n, err := drv.DelayedSize(context.Background(), "default")

	if err != nil {
		t.Fatal(err)
	}

	if n != 1 {
		t.Errorf("expected delayed size 1, got %d", n)
	}
}

func TestRedisDriverReservedSizeAlwaysZero(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	n, err := drv.ReservedSize(context.Background(), "default")

	if err != nil {
		t.Fatal(err)
	}

	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestRedisDriverConnectionName(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "my-redis")

	if drv.ConnectionName() != "my-redis" {
		t.Errorf("expected 'my-redis', got %q", drv.ConnectionName())
	}
}

func TestRedisDriverClearQueueDeletesQueueDelayedAndFailedKeys(t *testing.T) {
	t.Parallel()

	client := &deleterRedisClient{Client: redistest.NewClient()}
	drv := redis.NewDriver(client, "redis")

	if err := drv.ClearQueue(context.Background(), "emails"); err != nil {
		t.Fatalf("ClearQueue: %v", err)
	}

	want := []string{"queues:emails", "queues:emails:delayed", "queues:emails:failed"}

	if len(client.deleted) != len(want) {
		t.Fatalf("deleted %v, want %v", client.deleted, want)
	}

	for i, k := range want {
		if client.deleted[i] != k {
			t.Fatalf("deleted %v, want %v", client.deleted, want)
		}
	}
}

func TestRedisDriverClearQueueUsesDefaultForEmptyName(t *testing.T) {
	t.Parallel()

	client := &deleterRedisClient{Client: redistest.NewClient()}
	drv := redis.NewDriver(client, "redis")

	if err := drv.ClearQueue(context.Background(), ""); err != nil {
		t.Fatalf("ClearQueue: %v", err)
	}

	if len(client.deleted) == 0 || client.deleted[0] != "queues:default" {
		t.Fatalf("deleted %v, want the default queue key first", client.deleted)
	}
}

func TestRedisDriverClearQueuePropagatesDeleterError(t *testing.T) {
	t.Parallel()

	boom := errors.New("del failed")
	client := &deleterRedisClient{Client: redistest.NewClient(), delErr: boom}
	drv := redis.NewDriver(client, "redis")

	if err := drv.ClearQueue(context.Background(), "emails"); !errors.Is(err, boom) {
		t.Fatalf("ClearQueue error = %v, want %v", err, boom)
	}
}

// A client that cannot Del silently does nothing rather than erroring, because
// Deleter is an optional capability. Pin that: it is easy to "fix" into an
// ErrNotSupported and break callers that clear queues opportunistically.
func TestRedisDriverClearQueueWithoutDeleterIsSilentNoOp(t *testing.T) {
	t.Parallel()

	client := redistest.NewClient()
	drv := redis.NewDriver(client, "redis")

	if _, err := drv.Push(context.Background(), "emails", []byte(`{"job":"x"}`)); err != nil {
		t.Fatalf("Push: %v", err)
	}

	if err := drv.ClearQueue(context.Background(), "emails"); err != nil {
		t.Fatalf("ClearQueue without a deleter must be a no-op, got %v", err)
	}

	// The job must still be there — the no-op really is a no-op.
	size, err := drv.Size(context.Background(), "emails")

	if err != nil {
		t.Fatalf("Size: %v", err)
	}

	if size != 1 {
		t.Fatalf("size = %d, want 1 (ClearQueue must not have removed anything)", size)
	}
}

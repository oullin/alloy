package drivers_test

import (
	"context"
	"testing"
	"time"

	"github.com/oullin/alloy/pkg/hub/queue/drivers"
)

func TestRedisDriverPopMigratesDueDelayedJobs(t *testing.T) {
	t.Parallel()

	client := newMockRedisClient()
	drv := drivers.NewRedisDriver(client, "redis")

	_ = client.ZAdd(context.Background(), "queues:default:delayed", float64(time.Now().Add(-time.Minute).Unix()), "due-payload")

	job, err := drv.Pop(context.Background(), "default")

	if err != nil {
		t.Fatal(err)
	}

	if string(job.Payload()) != "due-payload" {
		t.Errorf("expected 'due-payload', got %q", job.Payload())
	}

	if len(client.evalCalls) != 1 {
		t.Fatalf("expected delayed migration to use one Lua Eval call, got %d", len(client.evalCalls))
	}

	if got := client.evalCalls[0].Keys; len(got) != 2 || got[0] != "queues:default:delayed" || got[1] != "queues:default" {
		t.Fatalf("unexpected Eval keys: %v", got)
	}

	n, _ := client.ZCard(context.Background(), "queues:default:delayed")

	if n != 0 {
		t.Errorf("expected 0 delayed after migration, got %d", n)
	}
}

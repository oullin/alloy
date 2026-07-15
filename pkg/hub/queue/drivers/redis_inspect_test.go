package drivers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oullin/alloy/pkg/hub/queue"
	"github.com/oullin/alloy/pkg/hub/queue/drivers"
)

// rangerRedisClient extends mockRedisClient with the optional Scanner /
// ListRanger / SortedSetRanger contracts used by the Redis driver's
// inspection methods.
type rangerRedisClient struct {
	*mockRedisClient
	keys      []string
	scanErr   error
	lrangeErr error
	zrangeErr error
}

func (c *rangerRedisClient) ScanMatch(_ context.Context, _ string) ([]string, error) {
	if c.scanErr != nil {
		return nil, c.scanErr
	}

	return c.keys, nil
}

func (c *rangerRedisClient) LRange(_ context.Context, key string, _, _ int64) ([]string, error) {
	if c.lrangeErr != nil {
		return nil, c.lrangeErr
	}

	return c.lists[key], nil
}

func (c *rangerRedisClient) ZRange(_ context.Context, key string, _, _ int64) ([]string, error) {
	if c.zrangeErr != nil {
		return nil, c.zrangeErr
	}

	entries := c.sorted[key]
	out := make([]string, 0, len(entries))

	for _, e := range entries {
		out = append(out, e.member)
	}

	return out, nil
}

func TestRedisDriverInspectionWithoutCapabilityReturnsErrNotSupported(t *testing.T) {
	t.Parallel()

	client := newMockRedisClient()
	drv := drivers.NewRedisDriver(client, "redis")
	ctx := context.Background()

	if _, err := drv.QueueNames(ctx); !errors.Is(err, queue.ErrNotSupported) {
		t.Errorf("QueueNames: want ErrNotSupported, got %v", err)
	}

	if _, err := drv.PendingJobs(ctx, "default"); !errors.Is(err, queue.ErrNotSupported) {
		t.Errorf("PendingJobs: want ErrNotSupported, got %v", err)
	}

	if _, err := drv.DelayedJobs(ctx, "default"); !errors.Is(err, queue.ErrNotSupported) {
		t.Errorf("DelayedJobs: want ErrNotSupported, got %v", err)
	}

	if _, err := drv.ReservedJobs(ctx, "default"); !errors.Is(err, queue.ErrNotSupported) {
		t.Errorf("ReservedJobs: want ErrNotSupported, got %v", err)
	}
}

func TestRedisDriverQueueNamesDedupesAndUnwraps(t *testing.T) {
	t.Parallel()

	client := &rangerRedisClient{
		mockRedisClient: newMockRedisClient(),
		keys: []string{
			"queues:default",
			"queues:default:delayed",
			"queues:emails",
			"queues:{cluster}",
			"queues:", // empty after trim — must be skipped
		},
	}
	drv := drivers.NewRedisDriver(client, "redis")

	names, err := drv.QueueNames(context.Background())

	if err != nil {
		t.Fatalf("QueueNames: %v", err)
	}

	want := map[string]struct{}{"default": {}, "emails": {}, "cluster": {}}

	if len(names) != len(want) {
		t.Fatalf("got %v, want %d unique names", names, len(want))
	}

	for _, n := range names {
		if _, ok := want[n]; !ok {
			t.Errorf("unexpected name %q in %v", n, names)
		}
	}
}

func TestRedisDriverQueueNamesScannerErrorPropagates(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("scan boom")
	client := &rangerRedisClient{
		mockRedisClient: newMockRedisClient(),
		scanErr:         wantErr,
	}
	drv := drivers.NewRedisDriver(client, "redis")

	_, err := drv.QueueNames(context.Background())

	if !errors.Is(err, wantErr) {
		t.Errorf("want %v, got %v", wantErr, err)
	}
}

func TestRedisDriverPendingJobsReturnsSnapshots(t *testing.T) {
	t.Parallel()

	client := &rangerRedisClient{mockRedisClient: newMockRedisClient()}
	drv := drivers.NewRedisDriver(client, "redis")
	ctx := context.Background()

	// Push two real payloads via the driver so the queue key matches.
	_, _ = drv.Push(ctx, "default", []byte(`{"uuid":"u1","displayName":"Job1","createdAt":1000000}`))
	_, _ = drv.Push(ctx, "default", []byte(`not-json`))

	jobs, err := drv.PendingJobs(ctx, "default")

	if err != nil {
		t.Fatalf("PendingJobs: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("got %d jobs, want 2", len(jobs))
	}

	// Find the JSON one and assert decoded fields.
	var decoded *queue.InspectedJob

	for i := range jobs {
		if jobs[i].UUID == "u1" {
			decoded = &jobs[i]

			break
		}
	}

	if decoded == nil {
		t.Fatal("expected a job with UUID u1")
	}

	if decoded.Name != "Job1" || decoded.Connection != "redis" || decoded.CreatedAt.Unix() != 1000000 {
		t.Errorf("decoded mismatch: %+v", decoded)
	}
}

func TestRedisDriverDelayedJobsReturnsSnapshots(t *testing.T) {
	t.Parallel()

	client := &rangerRedisClient{mockRedisClient: newMockRedisClient()}
	drv := drivers.NewRedisDriver(client, "redis")
	ctx := context.Background()

	_, _ = drv.PushDelayed(ctx, "default", []byte(`{"uuid":"d1"}`), 24*time.Hour)

	jobs, err := drv.DelayedJobs(ctx, "default")

	if err != nil {
		t.Fatalf("DelayedJobs: %v", err)
	}

	if len(jobs) != 1 || jobs[0].UUID != "d1" || jobs[0].Backend != "default" {
		t.Errorf("got %+v, want 1 job with UUID d1 on default", jobs)
	}
}

func TestRedisDriverInspectionRangerErrorPropagates(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("range boom")
	client := &rangerRedisClient{
		mockRedisClient: newMockRedisClient(),
		lrangeErr:       wantErr,
		zrangeErr:       wantErr,
	}
	drv := drivers.NewRedisDriver(client, "redis")
	ctx := context.Background()

	if _, err := drv.PendingJobs(ctx, "default"); !errors.Is(err, wantErr) {
		t.Errorf("PendingJobs: want %v, got %v", wantErr, err)
	}

	if _, err := drv.DelayedJobs(ctx, "default"); !errors.Is(err, wantErr) {
		t.Errorf("DelayedJobs: want %v, got %v", wantErr, err)
	}
}

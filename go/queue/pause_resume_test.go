package queue_test

import (
	"sync"
	"testing"
	"time"

	"alloy.dev/go/queue"
	"alloy.dev/go/queue/events"
)

// the upstream test constructs a QueueManager against a Carbon test clock
// and an ArrayStore-backed cache. The Go equivalent exercises the
// underlying PauseResumer + InMemoryPauseStore directly — when Step 6
// wires up the full Manager, the same contract must still hold and
// these tests continue to pass because Manager just delegates.
//
// Event assertions use the clockRecorder/pauseEventRecorder helpers below,
// which match Mockery's role in the PHP test (capture the dispatched
// event and assert on its fields).

// --- helpers ----------------------------------------------------------

// mockClock is a deterministic time source we can freeze and advance,
type mockClock struct {
	mu  sync.Mutex
	now time.Time
}

// pauseEventRecorder collects every event the PauseResumer emits so tests
// can assert the dispatched payloads.
type pauseEventRecorder struct {
	mu     sync.Mutex
	events []any
}

func newMockClock(now time.Time) *mockClock { return &mockClock{now: now} }

func (c *mockClock) Now() time.Time {
	c.mu.Lock()

	defer c.mu.Unlock()

	return c.now
}

func (c *mockClock) Advance(d time.Duration) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.now = c.now.Add(d)
}

func (r *pauseEventRecorder) Emit(event any) {
	r.mu.Lock()

	defer r.mu.Unlock()

	r.events = append(r.events, event)
}

func (r *pauseEventRecorder) last() any {
	r.mu.Lock()

	defer r.mu.Unlock()

	if len(r.events) == 0 {
		return nil
	}

	return r.events[len(r.events)-1]
}

// newPauseResumer wires up a PauseResumer with its own in-memory store,
// recorder, and mock clock. The returned clock is bound into the store
// so PauseFor tests can advance it deterministically.
func newPauseResumer() (*queue.PauseResumer, *pauseEventRecorder, *mockClock) {
	store := queue.NewInMemoryPauseStore()
	clock := newMockClock(time.Date(2026, 4, 14, 12, 0, 0, 0, time.UTC))
	store.SetClock(clock.Now)
	rec := &pauseEventRecorder{}

	return queue.NewPauseResumer(store, rec), rec, clock
}

// --- ports ------------------------------------------------------------

func TestPauseQueueWithConnection(t *testing.T) {
	t.Parallel()

	pr, _, _ := newPauseResumer()

	if err := pr.Pause("redis", "default"); err != nil {
		t.Fatalf("Pause: %v", err)
	}

	if !pr.IsPaused("redis", "default") {
		t.Error("expected redis:default to be paused")
	}
}

func TestPauseQueueWithTTL(t *testing.T) {
	t.Parallel()

	pr, _, clock := newPauseResumer()

	if err := pr.PauseFor("redis", "default", 30*time.Second); err != nil {
		t.Fatalf("PauseFor: %v", err)
	}

	if !pr.IsPaused("redis", "default") {
		t.Error("expected redis:default to be paused immediately after PauseFor")
	}

	clock.Advance(time.Minute)

	if pr.IsPaused("redis", "default") {
		t.Error("expected redis:default to have expired after advancing 1 minute")
	}
}

func TestPauseQueueIndefinitely(t *testing.T) {
	t.Parallel()

	pr, _, clock := newPauseResumer()

	if err := pr.Pause("redis", "default"); err != nil {
		t.Fatalf("Pause: %v", err)
	}

	if !pr.IsPaused("redis", "default") {
		t.Error("expected indefinite pause to be active immediately")
	}

	clock.Advance(365 * 24 * time.Hour)

	if !pr.IsPaused("redis", "default") {
		t.Error("expected indefinite pause to persist after 1 year")
	}
}

func TestResumeQueue(t *testing.T) {
	t.Parallel()

	pr, _, _ := newPauseResumer()

	_ = pr.Pause("redis", "default")

	if !pr.IsPaused("redis", "default") {
		t.Fatal("pre-condition: expected paused")
	}

	if err := pr.Resume("redis", "default"); err != nil {
		t.Fatalf("Resume: %v", err)
	}

	if pr.IsPaused("redis", "default") {
		t.Error("expected redis:default to be resumed")
	}
}

func TestPausingQueueOnOneConnectionDoesNotAffectAnother(t *testing.T) {
	t.Parallel()

	pr, _, _ := newPauseResumer()

	_ = pr.Pause("redis", "default")

	if !pr.IsPaused("redis", "default") {
		t.Error("expected redis:default paused")
	}

	if pr.IsPaused("database", "default") {
		t.Error("expected database:default NOT paused")
	}
}

func TestPausingDifferentQueuesOnSameConnection(t *testing.T) {
	t.Parallel()

	pr, _, _ := newPauseResumer()

	_ = pr.Pause("redis", "emails")
	_ = pr.Pause("redis", "notifications")

	if !pr.IsPaused("redis", "emails") {
		t.Error("expected redis:emails paused")
	}

	if !pr.IsPaused("redis", "notifications") {
		t.Error("expected redis:notifications paused")
	}

	if pr.IsPaused("redis", "default") {
		t.Error("expected redis:default NOT paused")
	}
}

func TestResumingOnlyAffectsSpecificQueue(t *testing.T) {
	t.Parallel()

	pr, _, _ := newPauseResumer()

	_ = pr.Pause("redis", "emails")
	_ = pr.Pause("redis", "notifications")

	_ = pr.Resume("redis", "emails")

	if pr.IsPaused("redis", "emails") {
		t.Error("expected redis:emails resumed")
	}

	if !pr.IsPaused("redis", "notifications") {
		t.Error("expected redis:notifications still paused")
	}
}

func TestPauseDispatchesQueuePausedEvent(t *testing.T) {
	t.Parallel()

	pr, rec, _ := newPauseResumer()

	_ = pr.Pause("redis", "default")

	ev, ok := rec.last().(events.Paused)

	if !ok {
		t.Fatalf("expected Paused event, got %T (%v)", rec.last(), rec.last())
	}

	if ev.ConnectionName != "redis" {
		t.Errorf("ConnectionName: got %q, want redis", ev.ConnectionName)
	}

	if ev.Backend != "default" {
		t.Errorf("Backend: got %q, want default", ev.Backend)
	}

	if ev.TTL != nil {
		t.Errorf("TTL: got %v, want nil", *ev.TTL)
	}
}

func TestPauseForDispatchesQueuePausedEventWithTTL(t *testing.T) {
	t.Parallel()

	pr, rec, _ := newPauseResumer()

	_ = pr.PauseFor("redis", "emails", 60*time.Second)

	ev, ok := rec.last().(events.Paused)

	if !ok {
		t.Fatalf("expected Paused event, got %T", rec.last())
	}

	if ev.ConnectionName != "redis" || ev.Backend != "emails" {
		t.Errorf("ConnectionName/Backend: got %q/%q", ev.ConnectionName, ev.Backend)
	}

	if ev.TTL == nil {
		t.Fatal("TTL: nil, want non-nil 60s")
	}

	if *ev.TTL != 60*time.Second {
		t.Errorf("TTL: got %s, want 60s", *ev.TTL)
	}
}

func TestResumeDispatchesQueueResumedEvent(t *testing.T) {
	t.Parallel()

	pr, rec, _ := newPauseResumer()

	_ = pr.Resume("database", "notifications")

	ev, ok := rec.last().(events.Resumed)

	if !ok {
		t.Fatalf("expected Resumed event, got %T", rec.last())
	}

	if ev.ConnectionName != "database" {
		t.Errorf("ConnectionName: got %q, want database", ev.ConnectionName)
	}

	if ev.Backend != "notifications" {
		t.Errorf("Backend: got %q, want notifications", ev.Backend)
	}
}

func TestParsingQueueString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw            string
		wantConnection string
		wantQueue      string
	}{
		{"", "redis", "default"},
		{"emails", "redis", "emails"},
		{"database:notifications", "database", "notifications"},
		{"redis:foo:bar", "redis", "foo:bar"},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.raw, func(t *testing.T) {
			t.Parallel()

			gotConn, gotQ := queue.ParseQueue(tc.raw, "redis")

			if gotConn != tc.wantConnection || gotQ != tc.wantQueue {
				t.Errorf("ParseQueue(%q): got (%q, %q), want (%q, %q)",
					tc.raw, gotConn, gotQ, tc.wantConnection, tc.wantQueue)
			}
		})
	}
}

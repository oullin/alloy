package session

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// gcProbe is a Handler whose GC blocks until released (or its context is
// cancelled) so tests can observe scheduling, concurrency, and shutdown.
type gcProbe struct {
	started   chan struct{}
	release   chan struct{}
	calls     atomic.Int32
	active    atomic.Int32
	maxActive atomic.Int32
	ctxDone   atomic.Bool
	panicNow  bool
}

func newGCProbe() *gcProbe {
	return &gcProbe{
		started: make(chan struct{}, 64),
		release: make(chan struct{}),
	}
}

func (p *gcProbe) Open(context.Context, string, string) error   { return nil }
func (p *gcProbe) Close(context.Context) error                  { return nil }
func (p *gcProbe) Read(context.Context, string) (string, error) { return "", nil }
func (p *gcProbe) Write(context.Context, string, string) error  { return nil }
func (p *gcProbe) Destroy(context.Context, string) error        { return nil }

func (p *gcProbe) GC(ctx context.Context, _ int) error {
	p.calls.Add(1)

	if p.panicNow {
		panic("gc boom")
	}

	n := p.active.Add(1)

	for {
		old := p.maxActive.Load()

		if n <= old || p.maxActive.CompareAndSwap(old, n) {
			break
		}
	}

	p.started <- struct{}{}

	select {
	case <-p.release:
	case <-ctx.Done():
		p.ctxDone.Store(true)
	}

	p.active.Add(-1)

	return nil
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		if cond() {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("condition not met within timeout")
}

func TestGCSchedulerRunsWithoutBlockingCaller(t *testing.T) {
	p := newGCProbe()
	g := newGCScheduler(context.Background(), p, 60)

	g.trigger() // must not block even though GC blocks

	select {
	case <-p.started:
	case <-time.After(2 * time.Second):
		t.Fatal("GC did not start in the background")
	}

	close(p.release)
	waitFor(t, func() bool { return g.running.Load() == false })

	if got := p.calls.Load(); got != 1 {
		t.Fatalf("expected exactly one GC call, got %d", got)
	}
}

func TestGCSchedulerIsSingleFlight(t *testing.T) {
	p := newGCProbe()
	g := newGCScheduler(context.Background(), p, 60)

	g.trigger()
	<-p.started // first sweep is now in flight and blocked

	// Fire a storm of triggers while the first sweep is still running.
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			g.trigger()
		}()
	}

	wg.Wait()

	// None of the storm triggers may have started a second concurrent sweep.
	if got := p.calls.Load(); got != 1 {
		t.Fatalf("single-flight violated: expected 1 GC call, got %d", got)
	}

	close(p.release)
	waitFor(t, func() bool { return g.running.Load() == false })

	if got := p.maxActive.Load(); got != 1 {
		t.Fatalf("single-flight violated: max concurrent sweeps was %d", got)
	}

	// The scheduler re-arms once the sweep finishes: a later trigger runs a
	// second sweep (release is already closed, so it completes immediately).
	g.trigger()
	waitFor(t, func() bool { return p.calls.Load() == 2 })
}

func TestGCSchedulerDoesNotStartAfterContextCancel(t *testing.T) {
	p := newGCProbe()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	g := newGCScheduler(ctx, p, 60)
	g.trigger()

	// Give any (erroneously spawned) goroutine a chance to run.
	time.Sleep(20 * time.Millisecond)

	if got := p.calls.Load(); got != 0 {
		t.Fatalf("no sweep must start once the lifecycle context is cancelled, got %d calls", got)
	}
}

func TestGCSchedulerInFlightSweepStopsOnCancel(t *testing.T) {
	p := newGCProbe() // release is never closed; only cancellation ends the sweep

	ctx, cancel := context.WithCancel(context.Background())
	g := newGCScheduler(ctx, p, 60)

	g.trigger()
	<-p.started

	cancel()

	waitFor(t, func() bool { return g.running.Load() == false })

	if !p.ctxDone.Load() {
		t.Fatal("in-flight sweep should observe context cancellation")
	}
}

func TestGCSchedulerRecoversFromPanic(t *testing.T) {
	p := newGCProbe()
	p.panicNow = true

	g := newGCScheduler(context.Background(), p, 60)

	g.trigger() // handler panics; the scheduler must recover and re-arm

	waitFor(t, func() bool { return p.calls.Load() == 1 && g.running.Load() == false })

	// A panic must not wedge the single-flight flag: a later trigger still runs.
	g.trigger()
	waitFor(t, func() bool { return p.calls.Load() == 2 })
}

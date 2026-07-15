package session

import (
	"context"
	"sync/atomic"
)

// gcScheduler runs session garbage collection off the request path.
//
// It is single-flight: at most one sweep runs at a time, and a trigger that
// arrives while a sweep is already in progress is dropped rather than queued,
// so a burst of probabilistic hits across many concurrent requests cannot spawn
// a storm of overlapping directory walks or bulk DELETEs. Each sweep runs in
// its own goroutine and recovers from panics so a misbehaving handler cannot
// crash the server. Sweeps are bound to a lifecycle context: once it is
// cancelled no new sweep starts, and the in-flight sweep observes the
// cancellation through the context it was handed.
type gcScheduler struct {
	handler     Handler
	ctx         context.Context
	maxLifetime int
	running     atomic.Bool
}

func newGCScheduler(ctx context.Context, handler Handler, maxLifetime int) *gcScheduler {
	if ctx == nil {
		ctx = context.Background()
	}

	return &gcScheduler{
		handler:     handler,
		ctx:         ctx,
		maxLifetime: maxLifetime,
	}
}

// trigger schedules a background sweep unless one is already running or the
// lifecycle context has ended. It never blocks the caller, so the request path
// never waits on GC.
func (g *gcScheduler) trigger() {
	if g.ctx.Err() != nil {
		return
	}

	// CompareAndSwap enforces single-flight: only the goroutine that flips the
	// flag from false to true starts a sweep; all others return immediately.
	if !g.running.CompareAndSwap(false, true) {
		return
	}

	go g.run()
}

func (g *gcScheduler) run() {
	// Reset the single-flight flag last so a panic still frees the next sweep,
	// and recover first so a handler panic never propagates past this goroutine.
	defer g.running.Store(false)

	defer func() { _ = recover() }()

	// The lifecycle context (not the request context) is handed to the handler
	// so an in-flight sweep is not aborted when the triggering request ends, yet
	// still stops promptly when the server shuts the lifecycle context down.
	_ = g.handler.GC(g.ctx, g.maxLifetime)
}

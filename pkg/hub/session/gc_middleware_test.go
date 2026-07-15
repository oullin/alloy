package session_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/oullin/alloy/pkg/hub/session"
	"github.com/oullin/alloy/pkg/hub/session/handlers"
)

// blockingGCHandler embeds ArrayHandler and blocks inside GC until released so
// the test can prove the request returned without waiting for GC.
type blockingGCHandler struct {
	*handlers.ArrayHandler
	started chan struct{}
	release chan struct{}
	gcCalls atomic.Int32
}

func newBlockingGCHandler() *blockingGCHandler {
	return &blockingGCHandler{
		ArrayHandler: handlers.NewArrayHandler(),
		started:      make(chan struct{}, 1),
		release:      make(chan struct{}),
	}
}

func (h *blockingGCHandler) GC(ctx context.Context, _ int) error {
	h.gcCalls.Add(1)
	h.started <- struct{}{}

	select {
	case <-h.release:
	case <-ctx.Done():
	}

	return nil
}

func TestStartSessionGCRunsOffRequestPath(t *testing.T) {
	h := newBlockingGCHandler()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw := session.StartSessionWithContext(ctx, h, session.StartSessionConfig{
		CookieName:    "sess",
		GCProbability: 100, // force a GC trigger on this request
	})

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	done := make(chan struct{})

	go func() {
		mw(inner).ServeHTTP(rr, req)
		close(done)
	}()

	// The request must complete even though GC is still blocked. If GC ran
	// inline, ServeHTTP would not return until we released it.
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		close(h.release)
		t.Fatal("request did not return; GC appears to run on the request path")
	}

	// GC was nonetheless scheduled in the background.
	select {
	case <-h.started:
	case <-time.After(2 * time.Second):
		t.Fatal("background GC was never scheduled")
	}

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Result().StatusCode)
	}

	close(h.release)
}

func TestStartSessionGCStopsOnContextCancel(t *testing.T) {
	h := newBlockingGCHandler()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // lifecycle already ended before any request

	mw := session.StartSessionWithContext(ctx, h, session.StartSessionConfig{
		CookieName:    "sess",
		GCProbability: 100,
	})

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(inner).ServeHTTP(rr, req)

	// No sweep may start once the lifecycle context is cancelled.
	time.Sleep(20 * time.Millisecond)

	if got := h.gcCalls.Load(); got != 0 {
		t.Fatalf("expected no GC after context cancel, got %d calls", got)
	}
}

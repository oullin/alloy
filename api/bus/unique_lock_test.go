package bus_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"alloy.dev/api/bus"
)

func TestUniqueLockAcquireSuccess(t *testing.T) {
	cache := newMockCacheStore()
	lock := bus.NewUniqueLock(cache)

	if !lock.Acquire(context.Background(), "my-job", 60*time.Second) {
		t.Error("expected Acquire to succeed on empty cache")
	}

	// Verify cache.Add was called with correct key and TTL.
	found := false

	for _, call := range cache.calls {
		if call.Method == "Add" && call.Key == "unique_lock:my-job" && call.TTL == 60 {
			found = true

			break
		}
	}

	if !found {
		t.Error("expected cache.Add to be called with key 'unique_lock:my-job' and TTL 60")
	}
}

func TestUniqueLockAcquireAlreadyHeld(t *testing.T) {
	cache := newMockCacheStore()
	cache.data["unique_lock:my-job"] = "1"

	lock := bus.NewUniqueLock(cache)

	if lock.Acquire(context.Background(), "my-job", 60*time.Second) {
		t.Error("expected Acquire to fail when lock is already held")
	}
}

func TestUniqueLockAcquireDefaultTTL(t *testing.T) {
	cache := newMockCacheStore()
	lock := bus.NewUniqueLock(cache)

	lock.Acquire(context.Background(), "key", 0)

	for _, call := range cache.calls {
		if call.Method == "Add" && call.TTL != 3600 {
			t.Errorf("expected default TTL 3600, got %d", call.TTL)
		}
	}
}

func TestUniqueLockRelease(t *testing.T) {
	cache := newMockCacheStore()
	cache.data["unique_lock:my-job"] = "1"

	lock := bus.NewUniqueLock(cache)

	if err := lock.Release(context.Background(), "my-job"); err != nil {
		t.Fatal(err)
	}

	found := false

	for _, call := range cache.calls {
		if call.Method == "Forget" && call.Key == "unique_lock:my-job" {
			found = true

			break
		}
	}

	if !found {
		t.Error("expected cache.Forget to be called with key 'unique_lock:my-job'")
	}
}

func TestUniqueLockKeyPrefix(t *testing.T) {
	cache := newMockCacheStore()
	lock := bus.NewUniqueLock(cache)

	lock.Acquire(context.Background(), "foo", 30*time.Second)

	for _, call := range cache.calls {
		if call.Method == "Add" && call.Key != "unique_lock:foo" {
			t.Errorf("expected key 'unique_lock:foo', got %q", call.Key)
		}
	}
}

func TestUniqueLockAcquireOnlyOneConcurrentAcquirer(t *testing.T) {
	cache := newMockCacheStore()
	lock := bus.NewUniqueLock(cache)

	var (
		wg        sync.WaitGroup
		successes int32
	)

	for range 20 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if lock.Acquire(context.Background(), "concurrent-job", 60*time.Second) {
				atomic.AddInt32(&successes, 1)
			}
		}()
	}

	wg.Wait()

	if successes != 1 {
		t.Fatalf("expected exactly one successful acquire, got %d", successes)
	}
}

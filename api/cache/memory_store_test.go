package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/oullin/alloy/api/cache"
)

func TestMemoryStorePutGetForgetAndExpiry(t *testing.T) {
	store := cache.NewMemoryStore()

	if err := store.Put(context.Background(), "key", "value", time.Minute); err != nil {
		t.Fatal(err)
	}

	value, ok, err := store.Get(context.Background(), "key")

	if err != nil {
		t.Fatal(err)
	}

	if !ok || value != "value" {
		t.Fatalf("value = %v, ok = %v", value, ok)
	}

	if err := store.Forget(context.Background(), "key"); err != nil {
		t.Fatal(err)
	}

	if _, ok, _ := store.Get(context.Background(), "key"); ok {
		t.Fatal("expected key to be forgotten")
	}
}

func TestMemoryStoreIncrementKeepsWindow(t *testing.T) {
	store := cache.NewMemoryStore()

	first, err := store.Increment(context.Background(), "hits", time.Minute)

	if err != nil {
		t.Fatal(err)
	}

	second, err := store.Increment(context.Background(), "hits", time.Minute)

	if err != nil {
		t.Fatal(err)
	}

	if first != 1 || second != 2 {
		t.Fatalf("hits = %d/%d", first, second)
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := cache.NewRateLimiter(cache.NewMemoryStore())

	tooMany, err := limiter.TooManyAttempts(context.Background(), "login", 2)

	if err != nil {
		t.Fatal(err)
	}

	if tooMany {
		t.Fatal("fresh key should not be throttled")
	}

	_, _ = limiter.Hit(context.Background(), "login", time.Minute)
	_, _ = limiter.Hit(context.Background(), "login", time.Minute)
	tooMany, err = limiter.TooManyAttempts(context.Background(), "login", 2)

	if err != nil {
		t.Fatal(err)
	}

	if !tooMany {
		t.Fatal("expected key to be throttled")
	}

	if err := limiter.Clear(context.Background(), "login"); err != nil {
		t.Fatal(err)
	}

	tooMany, err = limiter.TooManyAttempts(context.Background(), "login", 2)

	if err != nil {
		t.Fatal(err)
	}

	if tooMany {
		t.Fatal("cleared key should not be throttled")
	}
}

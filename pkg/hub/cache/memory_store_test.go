package cache_test

import (
	"context"
	"testing"
	"time"

	"hara.sh/alloy/cache"
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

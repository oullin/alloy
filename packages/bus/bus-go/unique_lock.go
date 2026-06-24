package bus

import (
	"context"
	"time"
)

// CacheStore is the minimal cache interface required by UniqueLock.
type CacheStore interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string, ttlSeconds int) error
	Add(ctx context.Context, key, value string, ttlSeconds int) (bool, error)
	Forget(ctx context.Context, key string) error
}

// UniqueLock provides distributed job deduplication using a cache backend.
type UniqueLock struct {
	cache CacheStore
}

// NewUniqueLock creates a UniqueLock backed by cache.
func NewUniqueLock(cache CacheStore) *UniqueLock {
	return &UniqueLock{cache: cache}
}

// Acquire attempts to acquire the lock for the given key. Returns true on success.
func (l *UniqueLock) Acquire(ctx context.Context, key string, ttl time.Duration) bool {
	ttlSeconds := int(ttl.Seconds())

	if ttlSeconds <= 0 {
		ttlSeconds = 3600
	}

	ok, err := l.cache.Add(ctx, l.key(key), "1", ttlSeconds)

	return err == nil && ok
}

// Release releases the lock for the given key.
func (l *UniqueLock) Release(ctx context.Context, key string) error {
	return l.cache.Forget(ctx, l.key(key))
}

func (l *UniqueLock) key(k string) string { return "unique_lock:" + k }

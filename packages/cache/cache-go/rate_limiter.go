package cache

import (
	"context"
	"time"
)

// RateLimiter provides cache-backed fixed-window throttling.
type RateLimiter struct {
	store Store
	now   func() time.Time
}

// NewRateLimiter creates a RateLimiter.
func NewRateLimiter(store Store) *RateLimiter {
	if store == nil {
		store = NewMemoryStore()
	}

	return &RateLimiter{store: store, now: time.Now}
}

// Hit increments a key in a fixed window.
func (l *RateLimiter) Hit(ctx context.Context, key string, window time.Duration) (int, error) {
	return l.store.Increment(ctx, key, window)
}

// TooManyAttempts reports whether the key has reached maxAttempts.
func (l *RateLimiter) TooManyAttempts(ctx context.Context, key string, maxAttempts int) (bool, error) {
	value, ok, err := l.store.Get(ctx, key)

	if err != nil || !ok {
		return false, err
	}

	count, ok := value.(int)

	if !ok {
		return false, nil
	}

	return count >= maxAttempts, nil
}

// Clear removes the rate-limit key.
func (l *RateLimiter) Clear(ctx context.Context, key string) error {
	return l.store.Forget(ctx, key)
}

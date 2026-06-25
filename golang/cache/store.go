package cache

import (
	"context"
	"errors"
	"time"
)

// Store is a TTL-aware cache contract.
type Store interface {
	Get(ctx context.Context, key string) (any, bool, error)
	Put(ctx context.Context, key string, value any, ttl time.Duration) error
	Forget(ctx context.Context, key string) error
	Increment(ctx context.Context, key string, ttl time.Duration) (int, error)
}

var ErrNotFound = errors.New("cache: key not found")

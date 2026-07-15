package redis

import "context"

// Client is the minimal Redis interface required by the Redis queue driver.
type Client interface {
	// LPush prepends values to a list.
	LPush(ctx context.Context, key string, values ...any) error
	// RPop removes and returns the last element of a list.
	RPop(ctx context.Context, key string) (string, error)
	// ZAdd adds a member with a score to a sorted set.
	ZAdd(ctx context.Context, key string, score float64, member string) error
	// ZRangeByScore returns members with scores in [min,max].
	ZRangeByScore(ctx context.Context, key string, min, max float64) ([]string, error)
	// ZRem removes members from a sorted set.
	ZRem(ctx context.Context, key string, members ...any) error
	// Eval runs a Lua script atomically.
	Eval(ctx context.Context, script string, keys []string, args ...any) (any, error)
	// LLen returns the length of a list.
	LLen(ctx context.Context, key string) (int64, error)
	// ZCard returns the cardinality of a sorted set.
	ZCard(ctx context.Context, key string) (int64, error)
}

// Deleter is the optional Redis capability needed by ClearQueue.
type Deleter interface {
	Del(ctx context.Context, keys ...string) (int64, error)
}

// Scanner is the optional Redis capability needed by QueueNames.
// A driver supports queue enumeration only when its underlying client
// can iterate keys matching a pattern — go-redis exposes this via
// the SCAN command, and most cluster-aware clients fan out a SCAN
// across nodes for the caller.
type Scanner interface {
	ScanMatch(ctx context.Context, match string) ([]string, error)
}

// ListRanger is the optional capability needed by PendingJobs to
// return the raw payloads currently waiting on the queue list. Without
// it the driver can still report a size (LLen) but cannot snapshot
// payloads.
type ListRanger interface {
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
}

// SortedSetRanger is the optional capability needed by DelayedJobs.
// It returns the members of a sorted set ordered by score, which
// matches the upstream ZRANGE semantics for the delayed-job set.
type SortedSetRanger interface {
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
}

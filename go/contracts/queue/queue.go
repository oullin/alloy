package queue

import (
	"context"
	"time"
)

// Backend defines the interface for a queue backend.
type Backend interface {
	Push(ctx context.Context, queue string, payload []byte) (string, error)
	PushDelayed(ctx context.Context, queue string, payload []byte, delay time.Duration) (string, error)
	PushMultiple(ctx context.Context, queue string, payloads [][]byte) ([]string, error)
	Pop(ctx context.Context, queue string) (Job, error)
	Size(ctx context.Context, queue string) (int64, error)
	PendingSize(ctx context.Context, queue string) (int64, error)
	DelayedSize(ctx context.Context, queue string) (int64, error)
	ReservedSize(ctx context.Context, queue string) (int64, error)
	ConnectionName() string
}

// Connector creates a Backend from a configuration map.
type Connector interface {
	Connect(config map[string]any) (Backend, error)
}

// Namer exposes a display name for queue-like values.
type Namer interface {
	QueueDisplayName() string
}

// AfterCommitMarker marks jobs that should dispatch after commit.
type AfterCommitMarker interface {
	QueueAfterCommit() bool
}

// BeforeCommitMarker marks jobs that should dispatch before commit.
type BeforeCommitMarker interface {
	QueueBeforeCommit() bool
}

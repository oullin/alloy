package events

import "time"

// Looping is dispatched once per iteration of the worker's main loop,
type Looping struct {
	ConnectionName string
	Backend        string
}

// Busy is dispatched by the queue:monitor command when a queue has
// more pending work than its configured threshold.
type Busy struct {
	ConnectionName string
	Backend        string
	Size           int64
}

// Paused is dispatched when an operator pauses a queue via the
// the upstream event carries an integer `ttl` (seconds, null for indefinite).
// The Go port uses a *time.Duration so callers can inspect both the
// nullability and the magnitude without unit conversion. A nil TTL means
// the pause is indefinite.
type Paused struct {
	ConnectionName string
	Backend        string
	TTL            *time.Duration
}

// Resumed is dispatched when an operator resumes a paused queue via
type Resumed struct {
	ConnectionName string
	Backend        string
}

// FailedOver is dispatched by the failover driver when it abandons
// one backend and switches to the next.
type FailedOver struct {
	From string
	To   string
	Err  error
}

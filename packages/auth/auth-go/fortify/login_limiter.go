package fortify

import (
	"context"
	"sync"
	"time"
)

type loginAttempt struct {
	count      int
	expiresAt  time.Time
	lockedAt   time.Time
	releasesAt time.Time
}

// MemoryLoginLimiter tracks login attempts in memory.
type MemoryLoginLimiter struct {
	mu       sync.Mutex
	max      int
	decay    time.Duration
	lockout  time.Duration
	attempts map[string]loginAttempt
}

// NewMemoryLoginLimiter creates a login limiter.
func NewMemoryLoginLimiter(max int, decay, lockout time.Duration) *MemoryLoginLimiter {
	if max <= 0 {
		max = 5
	}

	if decay <= 0 {
		decay = time.Minute
	}

	if lockout <= 0 {
		lockout = time.Minute
	}

	return &MemoryLoginLimiter{
		max:      max,
		decay:    decay,
		lockout:  lockout,
		attempts: make(map[string]loginAttempt),
	}
}

func (l *MemoryLoginLimiter) TooManyAttempts(_ context.Context, key string) bool {
	l.mu.Lock()

	defer l.mu.Unlock()

	attempt, ok := l.currentAttempt(key)

	if !ok {
		return false
	}

	return !attempt.releasesAt.IsZero() && attempt.releasesAt.After(time.Now())
}

func (l *MemoryLoginLimiter) Hit(_ context.Context, key string) error {
	l.mu.Lock()

	defer l.mu.Unlock()

	now := time.Now()
	attempt, ok := l.currentAttempt(key)

	if !ok {
		attempt = loginAttempt{expiresAt: now.Add(l.decay)}
	}

	attempt.count++

	if attempt.count >= l.max {
		attempt.lockedAt = now
		attempt.releasesAt = now.Add(l.lockout)
	}

	l.attempts[key] = attempt

	return nil
}

func (l *MemoryLoginLimiter) Clear(_ context.Context, key string) error {
	l.mu.Lock()

	defer l.mu.Unlock()

	delete(l.attempts, key)

	return nil
}

func (l *MemoryLoginLimiter) AvailableIn(_ context.Context, key string) time.Duration {
	l.mu.Lock()

	defer l.mu.Unlock()

	attempt, ok := l.currentAttempt(key)

	if !ok || attempt.releasesAt.IsZero() {
		return 0
	}

	return time.Until(attempt.releasesAt)
}

func (l *MemoryLoginLimiter) currentAttempt(key string) (loginAttempt, bool) {
	attempt, ok := l.attempts[key]

	if !ok {
		return loginAttempt{}, false
	}

	now := time.Now()

	if !attempt.releasesAt.IsZero() && !attempt.releasesAt.After(now) {
		delete(l.attempts, key)

		return loginAttempt{}, false
	}

	if attempt.expiresAt.Before(now) && attempt.releasesAt.Before(now) {
		delete(l.attempts, key)

		return loginAttempt{}, false
	}

	return attempt, true
}

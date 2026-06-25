package cache

import (
	"context"
	"sync"
	"time"
)

type entry struct {
	value     any
	expiresAt time.Time
}

// MemoryStore is an in-memory TTL cache suitable for tests and single-process apps.
type MemoryStore struct {
	mu    sync.Mutex
	items map[string]entry
	now   func() time.Time
}

// NewMemoryStore creates a MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]entry), now: time.Now}
}

func (s *MemoryStore) Get(_ context.Context, key string) (any, bool, error) {
	s.mu.Lock()

	defer s.mu.Unlock()

	item, ok := s.items[key]

	if !ok {
		return nil, false, nil
	}

	if expired(item, s.now()) {
		delete(s.items, key)

		return nil, false, nil
	}

	return item.value, true, nil
}

func (s *MemoryStore) Put(_ context.Context, key string, value any, ttl time.Duration) error {
	s.mu.Lock()

	defer s.mu.Unlock()

	s.items[key] = entry{value: value, expiresAt: expiresAt(s.now(), ttl)}

	return nil
}

func (s *MemoryStore) Forget(_ context.Context, key string) error {
	s.mu.Lock()

	defer s.mu.Unlock()

	delete(s.items, key)

	return nil
}

func (s *MemoryStore) Increment(_ context.Context, key string, ttl time.Duration) (int, error) {
	s.mu.Lock()

	defer s.mu.Unlock()

	now := s.now()
	item, ok := s.items[key]

	if !ok || expired(item, now) {
		s.items[key] = entry{value: 1, expiresAt: expiresAt(now, ttl)}

		return 1, nil
	}

	current, ok := item.value.(int)

	if !ok {
		current = 0
	}

	current++
	item.value = current
	s.items[key] = item

	return current, nil
}

func expiresAt(now time.Time, ttl time.Duration) time.Time {
	if ttl <= 0 {
		return time.Time{}
	}

	return now.Add(ttl)
}

func expired(item entry, now time.Time) bool {
	return !item.expiresAt.IsZero() && !item.expiresAt.After(now)
}

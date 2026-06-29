package passkeys

import (
	"context"
	"sync"

	"github.com/go-webauthn/webauthn/webauthn"
)

// SessionStore persists WebAuthn ceremony session data server-side.
type SessionStore interface {
	Put(ctx context.Context, key string, data webauthn.SessionData) error
	Get(ctx context.Context, key string) (webauthn.SessionData, error)
	Delete(ctx context.Context, key string) error
}

// MemorySessionStore stores WebAuthn sessions in memory.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]webauthn.SessionData
}

// NewMemorySessionStore creates a MemorySessionStore.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{sessions: make(map[string]webauthn.SessionData)}
}

func (s *MemorySessionStore) Put(_ context.Context, key string, data webauthn.SessionData) error {
	s.mu.Lock()

	defer s.mu.Unlock()

	s.sessions[key] = data

	return nil
}

func (s *MemorySessionStore) Get(_ context.Context, key string) (webauthn.SessionData, error) {
	s.mu.RLock()

	defer s.mu.RUnlock()

	data, ok := s.sessions[key]

	if !ok {
		return webauthn.SessionData{}, ErrCredentialNotFound
	}

	return data, nil
}

func (s *MemorySessionStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()

	defer s.mu.Unlock()

	delete(s.sessions, key)

	return nil
}

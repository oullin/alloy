package browserx

import (
	"context"
	"sync"
)

// MemoryRepository stores browser sessions in memory.
type MemoryRepository struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

// NewMemoryRepository creates a MemoryRepository.
func NewMemoryRepository(sessions ...Session) *MemoryRepository {
	repo := &MemoryRepository{sessions: make(map[string]Session)}

	for _, session := range sessions {
		repo.sessions[session.ID] = session
	}

	return repo
}

func (r *MemoryRepository) FindForUser(_ context.Context, userID string) ([]Session, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	sessions := make([]Session, 0)

	for _, session := range r.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

func (r *MemoryRepository) Revoke(_ context.Context, userID, sessionID string) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	session, ok := r.sessions[sessionID]

	if !ok || session.UserID != userID {
		return ErrSessionNotFound
	}

	delete(r.sessions, sessionID)

	return nil
}

func (r *MemoryRepository) RevokeOther(_ context.Context, userID, currentSessionID string) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	for id, session := range r.sessions {
		if session.UserID == userID && id != currentSessionID {
			delete(r.sessions, id)
		}
	}

	return nil
}

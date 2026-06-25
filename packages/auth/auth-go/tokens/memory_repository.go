package tokens

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// MemoryRepository stores personal access tokens in memory.
type MemoryRepository struct {
	mu     sync.RWMutex
	nextID int64
	tokens map[string]Token
}

// NewMemoryRepository creates an in-memory personal access token repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{tokens: make(map[string]Token)}
}

func (r *MemoryRepository) Create(_ context.Context, token Token) (Token, error) {
	r.mu.Lock()

	defer r.mu.Unlock()

	r.nextID++
	token.ID = strconv.FormatInt(r.nextID, 10)

	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}

	token.Abilities = append([]string(nil), token.Abilities...)
	r.tokens[token.ID] = cloneToken(token)

	return token, nil
}

func (r *MemoryRepository) Find(_ context.Context, id string) (Token, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	token, ok := r.tokens[id]

	if !ok {
		return Token{}, ErrTokenNotFound
	}

	return cloneToken(token), nil
}

func (r *MemoryRepository) FindForUser(_ context.Context, userID string) ([]Token, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	found := make([]Token, 0)

	for _, token := range r.tokens {
		if token.UserID == userID {
			found = append(found, cloneToken(token))
		}
	}

	return found, nil
}

func (r *MemoryRepository) Delete(_ context.Context, id, userID string) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	token, ok := r.tokens[id]

	if !ok || token.UserID != userID {
		return ErrTokenNotFound
	}

	delete(r.tokens, id)

	return nil
}

func (r *MemoryRepository) Revoke(_ context.Context, id, userID string) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	token, ok := r.tokens[id]

	if !ok || token.UserID != userID {
		return ErrTokenNotFound
	}

	now := time.Now()
	token.RevokedAt = &now
	r.tokens[id] = cloneToken(token)

	return nil
}

func (r *MemoryRepository) Touch(_ context.Context, id string) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	token, ok := r.tokens[id]

	if !ok {
		return ErrTokenNotFound
	}

	now := time.Now()
	token.LastUsedAt = &now
	r.tokens[id] = cloneToken(token)

	return nil
}

func cloneToken(token Token) Token {
	token.Abilities = append([]string(nil), token.Abilities...)
	token.LastUsedAt = cloneTime(token.LastUsedAt)
	token.ExpiresAt = cloneTime(token.ExpiresAt)
	token.RevokedAt = cloneTime(token.RevokedAt)

	return token
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	cloned := *value

	return &cloned
}

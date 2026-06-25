package passkeys

import (
	"bytes"
	"context"
	"sync"

	"github.com/go-webauthn/webauthn/webauthn"
)

// MemoryRepository stores passkey state in memory.
type MemoryRepository struct {
	mu          sync.RWMutex
	handles     map[string][]byte
	credentials map[string][]webauthn.Credential
}

// NewMemoryRepository creates a MemoryRepository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		handles:     make(map[string][]byte),
		credentials: make(map[string][]webauthn.Credential),
	}
}

func (r *MemoryRepository) GetOrCreateUserHandle(_ context.Context, userID string) ([]byte, error) {
	r.mu.Lock()

	defer r.mu.Unlock()

	if handle, ok := r.handles[userID]; ok {
		return append([]byte(nil), handle...), nil
	}

	handle, err := generateUserHandle()

	if err != nil {
		return nil, err
	}

	r.handles[userID] = append([]byte(nil), handle...)

	return handle, nil
}

func (r *MemoryRepository) UserIDByHandle(_ context.Context, handle []byte) (string, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	for userID, candidate := range r.handles {
		if bytes.Equal(candidate, handle) {
			return userID, nil
		}
	}

	return "", ErrUserHandleNotFound
}

func (r *MemoryRepository) CredentialsByUser(_ context.Context, userID string) ([]webauthn.Credential, error) {
	r.mu.RLock()

	defer r.mu.RUnlock()

	return append([]webauthn.Credential(nil), r.credentials[userID]...), nil
}

func (r *MemoryRepository) SaveCredential(_ context.Context, userID string, credential webauthn.Credential) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	r.credentials[userID] = append(r.credentials[userID], credential)

	return nil
}

func (r *MemoryRepository) UpdateCredential(_ context.Context, userID string, credential webauthn.Credential) error {
	r.mu.Lock()

	defer r.mu.Unlock()

	credentials := r.credentials[userID]

	for i, candidate := range credentials {
		if bytes.Equal(candidate.ID, credential.ID) {
			credentials[i] = credential
			r.credentials[userID] = credentials

			return nil
		}
	}

	return ErrCredentialNotFound
}

package passkeys

import (
	"context"
	"crypto/rand"
	"errors"

	"github.com/go-webauthn/webauthn/webauthn"
)

// Repository persists WebAuthn user handles and credentials.
type Repository interface {
	GetOrCreateUserHandle(ctx context.Context, userID string) ([]byte, error)
	UserIDByHandle(ctx context.Context, handle []byte) (string, error)
	CredentialsByUser(ctx context.Context, userID string) ([]webauthn.Credential, error)
	SaveCredential(ctx context.Context, userID string, credential webauthn.Credential) error
	UpdateCredential(ctx context.Context, userID string, credential webauthn.Credential) error
}

var (
	ErrCredentialNotFound = errors.New("passkeys: credential not found")
	ErrUserHandleNotFound = errors.New("passkeys: user handle not found")
)

func generateUserHandle() ([]byte, error) {
	handle := make([]byte, 32)

	if _, err := rand.Read(handle); err != nil {
		return nil, err
	}

	return handle, nil
}

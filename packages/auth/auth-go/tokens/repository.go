package tokens

import (
	"context"
	"errors"
	"time"
)

// ErrTokenNotFound is returned when a token cannot be found.

// ErrTokenInactive is returned when a token is expired or revoked.

// Repository persists personal access token records.
type Repository interface {
	Create(ctx context.Context, token Token) (Token, error)
	Find(ctx context.Context, id string) (Token, error)
	FindForUser(ctx context.Context, userID string) ([]Token, error)
	Delete(ctx context.Context, id, userID string) error
	Revoke(ctx context.Context, id, userID string) error
	Touch(ctx context.Context, id string) error
}

var (
	ErrTokenNotFound = errors.New("tokens: token not found")

	ErrTokenInactive = errors.New("tokens: token inactive")
)

// FindByPlainTextToken resolves a plaintext id|secret token against a repository.
func FindByPlainTextToken(ctx context.Context, repo Repository, plainText string) (Token, error) {
	id, secret, ok := parsePlainTextToken(plainText)

	if !ok {
		return Token{}, ErrTokenNotFound
	}

	token, err := repo.Find(ctx, id)

	if err != nil {
		return Token{}, err
	}

	if !MatchSecret(token.TokenHash, secret) {
		return Token{}, ErrTokenNotFound
	}

	if !token.Active(time.Now()) {
		return Token{}, ErrTokenInactive
	}

	return token, nil
}

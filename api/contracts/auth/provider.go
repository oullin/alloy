package contracts

import "context"

// UserProvider retrieves users from a persistence layer.
type UserProvider interface {
	RetrieveByID(ctx context.Context, id string) (User, error)
	RetrieveByToken(ctx context.Context, id string, token string) (User, error)
	RetrieveByCredentials(ctx context.Context, credentials map[string]string) (User, error)
	UpdateRememberToken(ctx context.Context, user User, token string) error
	ValidateCredentials(ctx context.Context, user User, credentials map[string]string) (bool, error)
	RehashPasswordIfRequired(ctx context.Context, user User, credentials map[string]string, force bool) error
}

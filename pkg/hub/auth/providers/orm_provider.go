package providers

import (
	"context"

	cauth "github.com/oullin/alloy/pkg/hub/contracts/auth"
)

// ModelQuery is the minimal interface for an ORM-backed user query.
// Callers inject their ORM's query builder implementing this interface.
type ModelQuery interface {
	// FindByID returns a user by primary key, or nil if not found.
	FindByID(ctx context.Context, id string) (cauth.User, error)
	// FindByToken returns a user matching id + rememberToken, or nil.
	FindByToken(ctx context.Context, id string, token string) (cauth.User, error)
	// FindByCredentials returns a user matching the given credentials (excluding password).
	FindByCredentials(ctx context.Context, credentials map[string]string) (cauth.User, error)
	// UpdateToken stores a new remember token for the given user.
	UpdateToken(ctx context.Context, user cauth.User, token string) error
	// UpdatePassword stores a new password hash for the given user.
	UpdatePassword(ctx context.Context, user cauth.User, passwordHash string) error
}

// ORMUserProvider retrieves users via an injected ORM ModelQuery interface.
type ORMUserProvider struct {
	model  ModelQuery
	hasher cauth.PasswordHasher
}

// NewORMUserProvider creates an ORMUserProvider.
func NewORMUserProvider(model ModelQuery, hasher cauth.PasswordHasher) *ORMUserProvider {
	return &ORMUserProvider{model: model, hasher: hasher}
}

func (p *ORMUserProvider) RetrieveByID(ctx context.Context, id string) (cauth.User, error) {
	return p.model.FindByID(ctx, id)
}

func (p *ORMUserProvider) RetrieveByToken(ctx context.Context, id string, token string) (cauth.User, error) {
	user, err := p.model.FindByToken(ctx, id, token)

	if err != nil || user == nil {
		return nil, err
	}

	if user.GetRememberToken() != token {
		return nil, nil
	}

	return user, nil
}

func (p *ORMUserProvider) UpdateRememberToken(ctx context.Context, user cauth.User, token string) error {
	return p.model.UpdateToken(ctx, user, token)
}

func (p *ORMUserProvider) RetrieveByCredentials(ctx context.Context, credentials map[string]string) (cauth.User, error) {
	// Strip password from query credentials.
	query := make(map[string]string, len(credentials))

	for k, v := range credentials {
		if k != "password" {
			query[k] = v
		}
	}

	return p.model.FindByCredentials(ctx, query)
}

func (p *ORMUserProvider) ValidateCredentials(ctx context.Context, user cauth.User, credentials map[string]string) (bool, error) {
	plain := credentials["password"]

	if plain == "" {
		return false, nil
	}

	return p.hasher.Check(ctx, plain, user.GetAuthPassword())
}

func (p *ORMUserProvider) RehashPasswordIfRequired(ctx context.Context, user cauth.User, credentials map[string]string, force bool) error {
	if !force && !p.hasher.NeedsRehash(user.GetAuthPassword()) {
		return nil
	}

	plain := credentials["password"]

	if plain == "" {
		return nil
	}

	hash, err := p.hasher.Hash(ctx, plain)

	if err != nil {
		return err
	}

	return p.model.UpdatePassword(ctx, user, hash)
}

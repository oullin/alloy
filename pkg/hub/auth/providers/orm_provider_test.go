package providers

import (
	"context"
	"testing"

	"hara.sh/alloy/auth/user"
	cauth "hara.sh/alloy/contracts/auth"
)

type ormQuery struct {
	updatedUser cauth.User
	updatedHash string
}

type ormHasher struct{}

func (q *ormQuery) FindByID(context.Context, string) (cauth.User, error) {
	return nil, nil
}

func (q *ormQuery) FindByToken(context.Context, string, string) (cauth.User, error) {
	return nil, nil
}

func (q *ormQuery) FindByCredentials(context.Context, map[string]string) (cauth.User, error) {
	return nil, nil
}

func (q *ormQuery) UpdateToken(context.Context, cauth.User, string) error {
	return nil
}

func (q *ormQuery) UpdatePassword(_ context.Context, user cauth.User, passwordHash string) error {
	q.updatedUser = user
	q.updatedHash = passwordHash

	return nil
}

func (ormHasher) Hash(context.Context, string) (string, error) {
	return "new-hash", nil
}

func (ormHasher) Check(context.Context, string, string) (bool, error) {
	return true, nil
}

func (ormHasher) NeedsRehash(string) bool {
	return true
}

func TestORMUserProviderRehashPasswordStoresNewHash(t *testing.T) {
	query := &ormQuery{}
	provider := NewORMUserProvider(query, ormHasher{})
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "old-hash"})

	err := provider.RehashPasswordIfRequired(context.Background(), user, map[string]string{
		"password": "secret",
	}, false)

	if err != nil {
		t.Fatal(err)
	}

	if query.updatedUser != user {
		t.Fatalf("updated user = %v, want %v", query.updatedUser, user)
	}

	if query.updatedHash != "new-hash" {
		t.Fatalf("updated hash = %q, want %q", query.updatedHash, "new-hash")
	}
}

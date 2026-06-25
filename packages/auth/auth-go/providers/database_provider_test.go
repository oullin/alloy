package providers

import (
	"context"
	"strings"
	"testing"

	"github.com/oullin/alloy/auth"
	cauth "github.com/oullin/alloy/auth/contracts/auth"
)

func TestDatabaseUserProviderRejectsUnsafeCredentialField(t *testing.T) {
	provider := NewDatabaseUserProvider(&providerDB{}, "users", auth.NewBcryptHasher(4), nil)

	_, err := provider.RetrieveByCredentials(context.Background(), map[string]string{
		"email; DROP TABLE users": "test@example.com",
	})

	if err == nil {
		t.Fatal("expected unsafe credential field to be rejected")
	}
}

func TestDatabaseUserProviderFallsBackWhenTableNameIsUnsafe(t *testing.T) {
	db := &providerDB{}
	provider := NewDatabaseUserProvider(db, "users; DROP TABLE users", auth.NewBcryptHasher(4), func(map[string]any) cauth.Authenticatable {
		return nil
	})

	_, _ = provider.RetrieveByID(context.Background(), "1")

	if !strings.HasPrefix(db.query, "SELECT * FROM users WHERE") {
		t.Fatalf("query = %q, want default users table", db.query)
	}
}

type providerDB struct {
	query string
}

func (db *providerDB) QueryRow(_ context.Context, query string, _ ...any) DBRow {
	db.query = query

	return providerRow{}
}

func (db *providerDB) Exec(_ context.Context, query string, _ ...any) error {
	db.query = query

	return nil
}

type providerRow struct{}

func (providerRow) Scan(_ ...any) error {
	return nil
}

package providers

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"hara.sh/alloy/auth/security"
	cauth "hara.sh/alloy/contracts/auth"
)

type providerDB struct {
	query string
	row   DBRow
}

type providerRow struct {
	err error
}

var errProviderScan = errors.New("scan failed")

func TestDatabaseUserProviderRejectsUnsafeCredentialField(t *testing.T) {
	provider := NewDatabaseUserProvider(&providerDB{}, "users", security.NewBcryptHasher(4), nil)

	_, err := provider.RetrieveByCredentials(context.Background(), map[string]string{
		"email; DROP TABLE users": "test@example.com",
	})

	if err == nil {
		t.Fatal("expected unsafe credential field to be rejected")
	}
}

func TestDatabaseUserProviderRejectsEmptyQueryableCredentials(t *testing.T) {
	db := &providerDB{}
	provider := NewDatabaseUserProvider(db, "users", security.NewBcryptHasher(4), nil)

	_, err := provider.RetrieveByCredentials(context.Background(), map[string]string{
		"password": "secret",
	})

	if err == nil {
		t.Fatal("expected empty queryable credentials to be rejected")
	}

	if db.query != "" {
		t.Fatalf("query = %q, want no query for empty credentials", db.query)
	}
}

func TestDatabaseUserProviderFallsBackWhenTableNameIsUnsafe(t *testing.T) {
	db := &providerDB{}
	provider := NewDatabaseUserProvider(db, "users; DROP TABLE users", security.NewBcryptHasher(4), func(map[string]any) cauth.User {
		return nil
	})

	_, _ = provider.RetrieveByID(context.Background(), "1")

	if !strings.HasPrefix(db.query, "SELECT * FROM users WHERE") {
		t.Fatalf("query = %q, want default users table", db.query)
	}
}

func TestDatabaseUserProviderTreatsNoRowsAsMissingUser(t *testing.T) {
	db := &providerDB{row: providerRow{err: sql.ErrNoRows}}
	provider := NewDatabaseUserProvider(db, "users", security.NewBcryptHasher(4), func(map[string]any) cauth.User {
		t.Fatal("row mapper should not be called for sql.ErrNoRows")

		return nil
	})

	user, err := provider.RetrieveByID(context.Background(), "1")

	if err != nil {
		t.Fatalf("RetrieveByID error = %v, want nil", err)
	}

	if user != nil {
		t.Fatalf("RetrieveByID user = %v, want nil", user)
	}
}

func TestDatabaseUserProviderReturnsScanErrors(t *testing.T) {
	db := &providerDB{row: providerRow{err: errProviderScan}}
	provider := NewDatabaseUserProvider(db, "users", security.NewBcryptHasher(4), func(map[string]any) cauth.User {
		t.Fatal("row mapper should not be called when scan fails")

		return nil
	})

	_, err := provider.RetrieveByID(context.Background(), "1")

	if !errors.Is(err, errProviderScan) {
		t.Fatalf("RetrieveByID error = %v, want %v", err, errProviderScan)
	}
}

func (db *providerDB) QueryRow(_ context.Context, query string, _ ...any) DBRow {
	db.query = query

	if db.row != nil {
		return db.row
	}

	return providerRow{}
}

func (db *providerDB) Exec(_ context.Context, query string, _ ...any) error {
	db.query = query

	return nil
}

func (r providerRow) Scan(_ ...any) error {
	return r.err
}

package passwords

import (
	"context"
	"testing"
	"time"
)

func TestMemoryRepositoryStoresTokenHash(t *testing.T) {
	repo := NewMemoryRepository(time.Hour)

	token, err := repo.Create(context.Background(), "test@example.com")

	if err != nil {
		t.Fatal(err)
	}

	entry := repo.tokens["test@example.com"]

	if entry.tokenHash == "" {
		t.Fatal("expected token hash to be stored")
	}

	if entry.tokenHash == token {
		t.Fatal("repository stored plaintext reset token")
	}

	if !repo.Exists(context.Background(), "test@example.com", token) {
		t.Fatal("repository should match the plaintext token against the stored hash")
	}
}

func TestSQLRepositoryStoresTokenHash(t *testing.T) {
	db := &tokenSQLDB{}
	repo := NewSQLRepository(db, "password_reset_tokens", time.Hour)

	token, err := repo.Create(context.Background(), "test@example.com")

	if err != nil {
		t.Fatal(err)
	}

	if db.token == "" {
		t.Fatal("expected token hash to be written")
	}

	if db.token == token {
		t.Fatal("SQL repository stored plaintext reset token")
	}

	if !repo.Exists(context.Background(), "test@example.com", token) {
		t.Fatal("repository should match plaintext token against stored hash")
	}
}

func TestSQLRepositoryFallsBackWhenTableNameIsUnsafe(t *testing.T) {
	db := &tokenSQLDB{}
	repo := NewSQLRepository(db, "password_reset_tokens; DROP TABLE users", time.Hour)

	_, err := repo.Create(context.Background(), "test@example.com")

	if err != nil {
		t.Fatal(err)
	}

	if db.query == "" {
		t.Fatal("expected insert query")
	}

	if want := "INSERT INTO password_reset_tokens "; len(db.query) < len(want) || db.query[:len(want)] != want {
		t.Fatalf("query = %q, want default password_reset_tokens table", db.query)
	}
}

type tokenSQLDB struct {
	query     string
	email     string
	token     string
	createdAt time.Time
}

func (db *tokenSQLDB) QueryRow(_ context.Context, query string, args ...any) SQLRow {
	db.query = query

	return tokenSQLRow{db: db}
}

func (db *tokenSQLDB) Exec(_ context.Context, query string, args ...any) error {
	db.query = query
	db.email, _ = args[0].(string)
	db.token, _ = args[1].(string)
	db.createdAt, _ = args[2].(time.Time)

	return nil
}

type tokenSQLRow struct {
	db *tokenSQLDB
}

func (r tokenSQLRow) Scan(dest ...any) error {
	if len(dest) == 2 {
		*dest[0].(*string) = r.db.token
		*dest[1].(*time.Time) = r.db.createdAt
	}

	return nil
}

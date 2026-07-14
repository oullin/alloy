package passwords

import (
	"context"
	"strings"
	"time"
)

// SQLQuerier is the minimal interface for SQL-backed password reset token storage.
type SQLQuerier interface {
	QueryRow(ctx context.Context, query string, args ...any) SQLRow
	Exec(ctx context.Context, query string, args ...any) error
}

// SQLRow is a single result row.
type SQLRow interface {
	Scan(dest ...any) error
}

// SQLRepository stores password reset token hashes in a SQL table with schema:
//
//	CREATE TABLE password_reset_tokens (
//	  email      TEXT PRIMARY KEY,
//	  token      TEXT NOT NULL,
//	  created_at TIMESTAMP NOT NULL
//	);
type SQLRepository struct {
	db     SQLQuerier
	table  string
	expiry time.Duration
}

// NewSQLRepository creates a SQLRepository.
func NewSQLRepository(db SQLQuerier, table string, expiry time.Duration) *SQLRepository {
	if !isSafeSQLIdentifier(table) {
		table = "password_reset_tokens"
	}

	return &SQLRepository{db: db, table: table, expiry: expiry}
}

func (r *SQLRepository) Create(ctx context.Context, email string) (string, error) {
	token, err := GenerateToken()

	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	tokenHash := hashToken(token)

	// Upsert: replace existing token if email already has one.
	err = r.db.Exec(ctx,
		"INSERT INTO "+r.table+" (email, token, created_at) VALUES ($1, $2, $3) "+
			"ON CONFLICT (email) DO UPDATE SET token=$2, created_at=$3",
		email, tokenHash, now,
	)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *SQLRepository) Exists(ctx context.Context, email, token string) bool {
	row := r.db.QueryRow(ctx,
		"SELECT token, created_at FROM "+r.table+" WHERE email = $1 LIMIT 1",
		email,
	)

	var storedToken string

	var createdAt time.Time

	if err := row.Scan(&storedToken, &createdAt); err != nil {
		return false
	}

	if time.Since(createdAt) > r.expiry {
		return false
	}

	return tokensMatch(storedToken, token)
}

func (r *SQLRepository) RecentlyCreated(ctx context.Context, email string, within time.Duration) bool {
	row := r.db.QueryRow(ctx,
		"SELECT created_at FROM "+r.table+" WHERE email = $1 LIMIT 1",
		email,
	)

	var createdAt time.Time

	if err := row.Scan(&createdAt); err != nil {
		return false
	}

	return time.Since(createdAt) <= within
}

func (r *SQLRepository) Delete(ctx context.Context, email string) error {
	return r.db.Exec(ctx, "DELETE FROM "+r.table+" WHERE email = $1", email)
}

func (r *SQLRepository) DeleteExpired(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-r.expiry)

	return r.db.Exec(ctx, "DELETE FROM "+r.table+" WHERE created_at < $1", cutoff)
}

func isSafeSQLIdentifier(identifier string) bool {
	if identifier == "" {
		return false
	}

	parts := strings.Split(identifier, ".")

	for _, part := range parts {
		if part == "" {
			return false
		}

		for i, r := range part {
			if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_' {
				continue
			}

			if i > 0 && r >= '0' && r <= '9' {
				continue
			}

			return false
		}
	}

	return true
}

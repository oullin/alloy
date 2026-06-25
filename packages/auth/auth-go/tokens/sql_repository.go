package tokens

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SQLQuerier is the minimal interface for SQL-backed personal access tokens.
type SQLQuerier interface {
	QueryRow(ctx context.Context, query string, args ...any) SQLRow
	Query(ctx context.Context, query string, args ...any) (SQLRows, error)
	Exec(ctx context.Context, query string, args ...any) error
}

// SQLRow is a single SQL result row.
type SQLRow interface {
	Scan(dest ...any) error
}

// SQLRows is a SQL result set.
type SQLRows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

// SQLRepository stores personal access token hashes in SQL.
type SQLRepository struct {
	db    SQLQuerier
	table string
	now   func() time.Time
}

// NewSQLRepository creates a SQL-backed personal access token repository.
func NewSQLRepository(db SQLQuerier, table string) *SQLRepository {
	if !isSafeSQLIdentifier(table) {
		table = "personal_access_tokens"
	}

	return &SQLRepository{
		db:    db,
		table: table,
		now:   time.Now,
	}
}

func (r *SQLRepository) Create(ctx context.Context, token Token) (Token, error) {
	if token.CreatedAt.IsZero() {
		token.CreatedAt = r.now()
	}

	abilities, err := json.Marshal(token.Abilities)

	if err != nil {
		return Token{}, err
	}

	row := r.db.QueryRow(ctx,
		"INSERT INTO "+r.table+" (user_id, name, token_hash, abilities, created_at, last_used_at, expires_at, revoked_at) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		token.UserID, token.Name, token.TokenHash, string(abilities), token.CreatedAt, token.LastUsedAt, token.ExpiresAt, token.RevokedAt,
	)

	if err := row.Scan(&token.ID); err != nil {
		return Token{}, err
	}

	return token, nil
}

func (r *SQLRepository) Find(ctx context.Context, id string) (Token, error) {
	row := r.db.QueryRow(ctx,
		"SELECT id, user_id, name, token_hash, abilities, created_at, last_used_at, expires_at, revoked_at FROM "+r.table+" WHERE id = $1 LIMIT 1",
		id,
	)

	return scanToken(row)
}

func (r *SQLRepository) FindForUser(ctx context.Context, userID string) ([]Token, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, user_id, name, token_hash, abilities, created_at, last_used_at, expires_at, revoked_at FROM "+r.table+" WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tokens := make([]Token, 0)

	for rows.Next() {
		token, err := scanToken(rows)

		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *SQLRepository) Delete(ctx context.Context, id, userID string) error {
	return r.db.Exec(ctx, "DELETE FROM "+r.table+" WHERE id = $1 AND user_id = $2", id, userID)
}

func (r *SQLRepository) Revoke(ctx context.Context, id, userID string) error {
	return r.db.Exec(ctx, "UPDATE "+r.table+" SET revoked_at = $1 WHERE id = $2 AND user_id = $3", r.now(), id, userID)
}

func (r *SQLRepository) Touch(ctx context.Context, id string) error {
	return r.db.Exec(ctx, "UPDATE "+r.table+" SET last_used_at = $1 WHERE id = $2", r.now(), id)
}

func scanToken(row interface{ Scan(dest ...any) error }) (Token, error) {
	var token Token

	var abilitiesJSON string

	if err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.Name,
		&token.TokenHash,
		&abilitiesJSON,
		&token.CreatedAt,
		&token.LastUsedAt,
		&token.ExpiresAt,
		&token.RevokedAt,
	); err != nil {
		return Token{}, err
	}

	if abilitiesJSON != "" {
		if err := json.Unmarshal([]byte(abilitiesJSON), &token.Abilities); err != nil {
			return Token{}, err
		}
	}

	return token, nil
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
			if i == 0 {
				if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && r != '_' {
					return false
				}

				continue
			}

			if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
				return false
			}
		}
	}

	return true
}

func (r *SQLRepository) String() string {
	return fmt.Sprintf("SQLRepository(%s)", r.table)
}

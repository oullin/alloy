package providers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	cauth "github.com/oullin/alloy/packages/foundation/contracts/auth"
)

// DBQuerier is the minimal raw-SQL interface for DatabaseUserProvider.
type DBQuerier interface {
	// QueryRow executes a query returning at most one row.
	QueryRow(ctx context.Context, query string, args ...any) DBRow
	// Exec executes a statement returning no rows.
	Exec(ctx context.Context, query string, args ...any) error
}

// DBRow is a single result row.
type DBRow interface {
	Scan(dest ...any) error
}

// RowMapper converts a scanned row map into an User.
type RowMapper func(row map[string]any) cauth.User

// DatabaseUserProvider retrieves users from a raw SQL table.
type DatabaseUserProvider struct {
	db        DBQuerier
	table     string
	hasher    cauth.PasswordHasher
	rowMapper RowMapper
}

// NewDatabaseUserProvider creates a DatabaseUserProvider.
// table is the users table name. rowMapper converts a row map to User.
func NewDatabaseUserProvider(db DBQuerier, table string, hasher cauth.PasswordHasher, rowMapper RowMapper) *DatabaseUserProvider {
	if !isSafeIdentifier(table) {
		table = "users"
	}

	return &DatabaseUserProvider{
		db:        db,
		table:     table,
		hasher:    hasher,
		rowMapper: rowMapper,
	}
}

func (p *DatabaseUserProvider) RetrieveByID(ctx context.Context, id string) (cauth.User, error) {
	row := p.db.QueryRow(ctx, fmt.Sprintf("SELECT * FROM %s WHERE id = $1 LIMIT 1", p.table), id)

	return p.mapRow(row)
}

func (p *DatabaseUserProvider) RetrieveByToken(ctx context.Context, id string, token string) (cauth.User, error) {
	row := p.db.QueryRow(ctx,
		fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND remember_token = $2 LIMIT 1", p.table),
		id, token,
	)

	return p.mapRow(row)
}

func (p *DatabaseUserProvider) UpdateRememberToken(ctx context.Context, user cauth.User, token string) error {
	return p.db.Exec(ctx,
		fmt.Sprintf("UPDATE %s SET remember_token = $1 WHERE id = $2", p.table),
		token, user.GetAuthIdentifier(),
	)
}

func (p *DatabaseUserProvider) RetrieveByCredentials(ctx context.Context, credentials map[string]string) (cauth.User, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE ", p.table)

	args := make([]any, 0, len(credentials))
	i := 1

	for k, v := range credentials {
		if k == "password" {
			continue
		}

		if !isSafeIdentifier(k) {
			return nil, fmt.Errorf("auth: unsafe credential field %q", k)
		}

		if i > 1 {
			query += " AND "
		}

		query += fmt.Sprintf("%s = $%d", k, i)
		args = append(args, v)
		i++
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("auth: credentials must include at least one queryable field")
	}

	query += " LIMIT 1"
	row := p.db.QueryRow(ctx, query, args...)

	return p.mapRow(row)
}

func (p *DatabaseUserProvider) ValidateCredentials(ctx context.Context, user cauth.User, credentials map[string]string) (bool, error) {
	plain := credentials["password"]

	if plain == "" {
		return false, nil
	}

	return p.hasher.Check(ctx, plain, user.GetAuthPassword())
}

func isSafeIdentifier(identifier string) bool {
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

func (p *DatabaseUserProvider) RehashPasswordIfRequired(ctx context.Context, user cauth.User, credentials map[string]string, force bool) error {
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

	return p.db.Exec(ctx,
		fmt.Sprintf("UPDATE %s SET password = $1 WHERE id = $2", p.table),
		hash, user.GetAuthIdentifier(),
	)
}

func (p *DatabaseUserProvider) mapRow(row DBRow) (cauth.User, error) {
	// Scan into a map via column names is not directly supported by the DBRow
	// interface; callers must provide a RowMapper that matches their DB driver's
	// row type. Here we delegate to the injected mapper.
	//
	// When the row mapper cannot be called directly (scan requires concrete types),
	// callers should embed the concrete row type and cast appropriately. This
	// interface-based approach matches the driver-agnostic design of this package.
	//
	// For a concrete example using database/sql, wrap *sql.Row so Scan returns
	// column values into a []any and then build the map before calling rowMapper.
	var dest map[string]any

	if err := row.Scan(&dest); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return p.rowMapper(dest), nil
}

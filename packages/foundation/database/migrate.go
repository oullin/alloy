package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Migration defines a versioned database schema change.
type Migration struct {
	Version int
	Name    string
	Up      func(ctx context.Context, tx *sql.Tx) error
}

// StatementsMigration creates a migration from SQL statements.
func StatementsMigration(version int, name string, stmts ...string) Migration {
	return Migration{
		Version: version,
		Name:    name,
		Up: func(ctx context.Context, tx *sql.Tx) error {
			for _, stmt := range stmts {
				if _, err := tx.ExecContext(ctx, stmt); err != nil {
					return fmt.Errorf("database: execute migration statement: %w", err)
				}
			}

			return nil
		},
	}
}

// Migrate applies unapplied migrations in ascending version order.
func Migrate(ctx context.Context, db *sql.DB, migrations []Migration) error {
	if err := validateMigrations(migrations); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, name TEXT NOT NULL, applied_at TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("database: create schema_migrations: %w", err)
	}

	applied, err := appliedMigrations(ctx, db)

	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		err := WithTx(ctx, db, nil, func(tx *sql.Tx) error {
			if migration.Up != nil {
				if err := migration.Up(ctx, tx); err != nil {
					return fmt.Errorf("database: apply migration %d %s: %w", migration.Version, migration.Name, err)
				}
			}

			appliedAt := time.Now().UTC().Format(time.RFC3339)
			query := fmt.Sprintf(
				"INSERT INTO schema_migrations (version, name, applied_at) VALUES (%d, '%s', '%s')",
				migration.Version,
				sqlString(migration.Name),
				sqlString(appliedAt),
			)

			if _, err := tx.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("database: record migration %d %s: %w", migration.Version, migration.Name, err)
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func validateMigrations(migrations []Migration) error {
	seen := make(map[int]struct{}, len(migrations))
	previous := 0

	for i, migration := range migrations {
		if migration.Version <= 0 {
			return fmt.Errorf("database: migration at index %d has non-positive version %d", i, migration.Version)
		}

		if _, ok := seen[migration.Version]; ok {
			return fmt.Errorf("database: duplicate migration version %d", migration.Version)
		}

		if migration.Version < previous {
			return fmt.Errorf("database: migration version %d descends after version %d", migration.Version, previous)
		}

		seen[migration.Version] = struct{}{}
		previous = migration.Version
	}

	return nil
}

func appliedMigrations(ctx context.Context, db *sql.DB) (map[int]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")

	if err != nil {
		return nil, fmt.Errorf("database: read schema_migrations: %w", err)
	}

	defer rows.Close()

	applied := make(map[int]bool)

	for rows.Next() {
		var version int

		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("database: scan schema_migrations: %w", err)
		}

		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database: read schema_migrations rows: %w", err)
	}

	return applied, nil
}

func sqlString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

package database_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	dbpkg "alloy.dev/foundation/database"
)

func TestMigrateAppliesAllAndRecordsVersions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openTestDB(t)
	applied := 0
	migrations := []dbpkg.Migration{
		dbpkg.StatementsMigration(1, "create widgets", "CREATE TABLE widgets (id INTEGER PRIMARY KEY, name TEXT NOT NULL)"),
		{
			Version: 2,
			Name:    "seed widgets",
			Up: func(ctx context.Context, tx *sql.Tx) error {
				applied++
				_, err := tx.ExecContext(ctx, "INSERT INTO widgets (name) VALUES ('alpha')")

				return err
			},
		},
	}

	if err := dbpkg.Migrate(ctx, db, migrations); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	if got := countRows(t, db, "widgets"); got != 1 {
		t.Fatalf("expected widget row count 1, got %d", got)
	}

	if got := countRows(t, db, "schema_migrations"); got != 2 {
		t.Fatalf("expected migration row count 2, got %d", got)
	}

	assertMigrationVersions(t, db, []int{1, 2})

	if err := dbpkg.Migrate(ctx, db, migrations); err != nil {
		t.Fatalf("second Migrate returned error: %v", err)
	}

	if applied != 1 {
		t.Fatalf("expected second run to skip applied migration, applied count %d", applied)
	}

	if got := countRows(t, db, "widgets"); got != 1 {
		t.Fatalf("expected widget row count to remain 1, got %d", got)
	}
}

func TestMigrateRollsBackFailingMigrationOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openTestDB(t)
	failErr := errors.New("migration failed")
	migrations := []dbpkg.Migration{
		dbpkg.StatementsMigration(1, "create widgets", "CREATE TABLE widgets (id INTEGER PRIMARY KEY, name TEXT NOT NULL)"),
		{
			Version: 2,
			Name:    "partial seed",
			Up: func(ctx context.Context, tx *sql.Tx) error {
				if _, err := tx.ExecContext(ctx, "INSERT INTO widgets (name) VALUES ('partial')"); err != nil {
					return err
				}

				return failErr
			},
		},
	}

	err := dbpkg.Migrate(ctx, db, migrations)
	if !errors.Is(err, failErr) {
		t.Fatalf("expected failing migration error, got %v", err)
	}

	if got := countRows(t, db, "widgets"); got != 0 {
		t.Fatalf("expected failing migration changes rolled back, got %d rows", got)
	}

	assertMigrationVersions(t, db, []int{1})
}

func TestMigrateRejectsInvalidVersionsBeforeApplying(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name       string
		migrations []dbpkg.Migration
		want       string
	}{
		{
			name:       "non-positive",
			migrations: []dbpkg.Migration{{Version: 0, Name: "zero"}},
			want:       "non-positive version 0",
		},
		{
			name:       "duplicate",
			migrations: []dbpkg.Migration{{Version: 1, Name: "one"}, {Version: 1, Name: "one again"}},
			want:       "duplicate migration version 1",
		},
		{
			name:       "descending",
			migrations: []dbpkg.Migration{{Version: 2, Name: "two"}, {Version: 1, Name: "one"}},
			want:       "version 1 descends after version 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := openTestDB(t)
			err := dbpkg.Migrate(ctx, db, tt.migrations)

			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}

			if tableExists(t, db, "schema_migrations") {
				t.Fatal("schema_migrations table was created after validation failure")
			}
		})
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close sqlite: %v", err)
		}
	})

	return db
}

func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatalf("count rows from %s: %v", table, err)
	}

	return count
}

func assertMigrationVersions(t *testing.T, db *sql.DB, want []int) {
	t.Helper()

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	defer rows.Close()

	var got []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			t.Fatalf("scan migration version: %v", err)
		}
		got = append(got, version)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("read migration versions: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("expected migration versions %v, got %v", want, got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected migration versions %v, got %v", want, got)
		}
	}
}

func tableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()

	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&name)

	if errors.Is(err, sql.ErrNoRows) {
		return false
	}

	if err != nil {
		t.Fatalf("check table exists: %v", err)
	}

	return true
}

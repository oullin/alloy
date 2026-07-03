package database_test

import (
	"testing"
	"testing/fstest"

	"github.com/golang-migrate/migrate/v4/database/sqlite"
	dbpkg "github.com/oullin/alloy/packages/foundation/database"
)

func migrationFS(files map[string]string) fstest.MapFS {
	fsys := make(fstest.MapFS, len(files))

	for name, body := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(body)}
	}

	return fsys
}

func TestMigrateAppliesPendingMigrations(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		t.Fatalf("sqlite driver: %v", err)
	}

	fsys := migrationFS(map[string]string{
		"migrations/1_create_widgets.up.sql": "CREATE TABLE widgets (id INTEGER PRIMARY KEY, name TEXT NOT NULL);",
		"migrations/2_seed_widget.up.sql":    "INSERT INTO widgets (name) VALUES ('gadget');",
	})

	if err := dbpkg.Migrate(fsys, "migrations", driver, "sqlite"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if got := countRows(t, db, "widgets"); got != 1 {
		t.Fatalf("widgets = %d, want 1", got)
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	fsys := migrationFS(map[string]string{
		"migrations/1_create_widgets.up.sql": "CREATE TABLE widgets (id INTEGER PRIMARY KEY, name TEXT NOT NULL);",
	})

	// Running the same set twice must not error and must leave the schema
	// untouched on the second pass.
	for range 2 {
		driver, err := sqlite.WithInstance(db, &sqlite.Config{})
		if err != nil {
			t.Fatalf("sqlite driver: %v", err)
		}

		if err := dbpkg.Migrate(fsys, "migrations", driver, "sqlite"); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}

	if got := countRows(t, db, "widgets"); got != 0 {
		t.Fatalf("widgets = %d, want 0", got)
	}
}

func TestMigrateReportsBrokenStatement(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		t.Fatalf("sqlite driver: %v", err)
	}

	fsys := migrationFS(map[string]string{
		"migrations/1_broken.up.sql": "CREATE TABLE (this is not valid sql;",
	})

	if err := dbpkg.Migrate(fsys, "migrations", driver, "sqlite"); err == nil {
		t.Fatal("expected error for invalid migration statement")
	}
}

func TestMigrateErrorsOnMissingDirectory(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		t.Fatalf("sqlite driver: %v", err)
	}

	if err := dbpkg.Migrate(fstest.MapFS{}, "migrations", driver, "sqlite"); err == nil {
		t.Fatal("expected error for missing migrations directory")
	}
}

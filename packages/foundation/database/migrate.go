package database

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	migratedb "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrate applies every pending "up" migration found in sourceFS under dir,
// using golang-migrate.
//
// Migration files follow golang-migrate's naming convention,
// {version}_{title}.up.sql (with matching .down.sql for rollbacks). sourceFS
// is any fs.FS — typically an embed.FS shipped with the application.
//
// The database driver is supplied by the caller (for example
// sqlite.WithInstance or postgres.WithInstance) so this package stays
// database-agnostic and pulls in no specific database dependency;
// databaseName must match that driver ("sqlite", "postgres", …).
//
// Migrate does not close the provided driver or its underlying *sql.DB — the
// caller owns that lifecycle. It returns nil when the schema is already up to
// date.
func Migrate(sourceFS fs.FS, dir string, driver migratedb.Driver, databaseName string) error {
	source, err := iofs.New(sourceFS, dir)

	if err != nil {
		return fmt.Errorf("database: open migration source: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", source, databaseName, driver)

	if err != nil {
		return fmt.Errorf("database: create migrator: %w", err)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("database: apply migrations: %w", err)
	}

	return nil
}

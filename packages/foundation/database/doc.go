// Package database contains shared database errors, a transaction helper
// (WithTx), and a thin golang-migrate wrapper (Migrate) for applying
// versioned SQL migrations. The migration helper is database-agnostic: the
// caller supplies the golang-migrate driver for their database, so this
// package pulls in no specific database dependency.
package database

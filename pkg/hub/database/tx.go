package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Transactor begins SQL transactions.
type Transactor interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// WithTx executes fn inside a SQL transaction.
func WithTx(ctx context.Context, db Transactor, opts *sql.TxOptions, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, opts)

	if err != nil {
		return fmt.Errorf("database: begin tx: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()

			panic(r)
		}
	}()

	if err = fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(err, fmt.Errorf("database: rollback tx: %w", rollbackErr))
		}

		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("database: commit tx: %w", err)
	}

	return nil
}

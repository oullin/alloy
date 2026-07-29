package database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	dbpkg "hara.sh/alloy/database"
)

func TestWithTxCommitsOnSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openTestDB(t)

	_, err := db.ExecContext(ctx, "CREATE TABLE entries (name TEXT NOT NULL)")

	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	err = dbpkg.WithTx(ctx, db, nil, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO entries (name) VALUES ('committed')")

		return err
	})

	if err != nil {
		t.Fatalf("WithTx returned error: %v", err)
	}

	if got := countRows(t, db, "entries"); got != 1 {
		t.Fatalf("expected committed row count 1, got %d", got)
	}
}

func TestWithTxRollsBackOnError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openTestDB(t)
	wantErr := errors.New("callback failed")

	_, err := db.ExecContext(ctx, "CREATE TABLE entries (name TEXT NOT NULL)")

	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	err = dbpkg.WithTx(ctx, db, nil, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "INSERT INTO entries (name) VALUES ('rolled back')"); err != nil {
			return err
		}

		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected callback error, got %v", err)
	}

	if got := countRows(t, db, "entries"); got != 0 {
		t.Fatalf("expected rolled back row count 0, got %d", got)
	}
}

func TestWithTxRollsBackAndRepanics(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openTestDB(t)
	panicValue := "callback panicked"

	_, err := db.ExecContext(ctx, "CREATE TABLE entries (name TEXT NOT NULL)")

	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	func() {
		defer func() {
			if r := recover(); r != panicValue {
				t.Fatalf("expected panic %q, got %v", panicValue, r)
			}
		}()

		_ = dbpkg.WithTx(ctx, db, nil, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, "INSERT INTO entries (name) VALUES ('rolled back')"); err != nil {
				return err
			}

			panic(panicValue)
		})
	}()

	if got := countRows(t, db, "entries"); got != 0 {
		t.Fatalf("expected rolled back row count 0, got %d", got)
	}
}

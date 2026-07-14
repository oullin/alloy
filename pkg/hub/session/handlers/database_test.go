package handlers_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/oullin/alloy/pkg/hub/session/handlers"
)

type mockDBConn struct {
	row        *mockDBRow
	query      string
	queryArgs  []any
	execQuery  string
	execArgs   []any
	execErr    error
	queryCalls int
	execCalls  int
}

type mockDBRow struct {
	payload string
	err     error
}

func (db *mockDBConn) QueryRow(_ context.Context, query string, args ...any) handlers.DBRow {
	db.query = query
	db.queryArgs = args
	db.queryCalls++

	if db.row == nil {
		return &mockDBRow{err: sql.ErrNoRows}
	}

	return db.row
}

func (db *mockDBConn) Exec(_ context.Context, query string, args ...any) error {
	db.execQuery = query
	db.execArgs = args
	db.execCalls++

	return db.execErr
}

func (r *mockDBRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	*(dest[0].(*string)) = r.payload

	return nil
}

func TestDatabaseHandlerReadReturnsEmptyForNoRows(t *testing.T) {
	db := &mockDBConn{row: &mockDBRow{err: sql.ErrNoRows}}
	h := handlers.NewDatabaseHandler(db, "sessions")

	got, err := h.Read(context.Background(), "missing")

	if err != nil {
		t.Fatal(err)
	}

	if got != "" {
		t.Fatalf("expected empty session, got %q", got)
	}
}

func TestDatabaseHandlerReadPropagatesDatabaseErrors(t *testing.T) {
	wantErr := errors.New("connection failed")
	db := &mockDBConn{row: &mockDBRow{err: wantErr}}
	h := handlers.NewDatabaseHandler(db, "sessions")

	_, err := h.Read(context.Background(), "session-id")

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected database error, got %v", err)
	}
}

func TestDatabaseHandlerSanitizesInvalidTableName(t *testing.T) {
	db := &mockDBConn{row: &mockDBRow{payload: "payload"}}
	h := handlers.NewDatabaseHandler(db, "sessions; DROP TABLE users")

	if _, err := h.Read(context.Background(), "session-id"); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(db.query, "DROP") || strings.Contains(db.query, ";") {
		t.Fatalf("query contains unsafe table name: %q", db.query)
	}

	if !strings.Contains(db.query, "FROM sessions WHERE") {
		t.Fatalf("expected fallback sessions table, got %q", db.query)
	}
}

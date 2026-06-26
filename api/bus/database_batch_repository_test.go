package bus_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/oullin/alloy/api/bus"
)

// mockDBExecutor implements bus.DBExecutor for unit testing.
type mockDBExecutor struct {
	mu          sync.Mutex
	execCalls   []mockDBExecCall
	execErr     error
	execResult  sql.Result
	queryRow    *mockRow
	queryRowErr error
	queryRows   *mockRows
	queryErr    error
}

type mockDBExecCall struct {
	Query string
	Args  []any
}

type mockResult struct {
	lastID       int64
	rowsAffected int64
}

// mockRow provides a scannable row for testing.
type mockRow struct {
	values []any
	err    error
}

// mockRows provides scannable rows for testing.
type mockRows struct {
	rows    [][]any
	cursor  int
	columns []string
}

// failed_job_ids is arg index 5.

// options is arg index 6.

// First call returns 1000 rows affected, second returns 0.

// The query should reference job_batches.

// mockDynamicResult lets RowsAffected return different values per call.
type mockDynamicResult struct {
	fn func() int64
}

type lockingBatchDriver struct{}

type lockingBatchConn struct {
	state *lockingBatchDriverState
}

type lockingBatchTx struct {
	state *lockingBatchDriverState
}

type lockingBatchRows struct {
	columns []string
	values  []driver.Value
	read    bool
}

type lockingBatchStmt struct{}

type lockingBatchDriverState struct {
	mu        sync.Mutex
	pending   int
	failed    int
	failedIDs []string
	queries   []string
	commits   int
	rollbacks int
}

var sqlLockingState = &lockingBatchDriverState{}

func init() {
	sql.Register("alloy-bus-locking-test", lockingBatchDriver{})
}

func newMockDBExecutor() *mockDBExecutor {
	return &mockDBExecutor{
		execResult: &mockResult{rowsAffected: 1},
	}
}

func (db *mockDBExecutor) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	db.mu.Lock()

	defer db.mu.Unlock()

	db.execCalls = append(db.execCalls, mockDBExecCall{Query: query, Args: args})

	return db.execResult, db.execErr
}

func (db *mockDBExecutor) QueryContext(_ context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, db.queryErr
}

func (db *mockDBExecutor) QueryRowContext(_ context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (r *mockResult) LastInsertId() (int64, error) { return r.lastID, nil }
func (r *mockResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

func (d lockingBatchDriver) Open(_ string) (driver.Conn, error) {
	return &lockingBatchConn{state: sqlLockingState}, nil
}

func (c *lockingBatchConn) Prepare(_ string) (driver.Stmt, error) {
	return lockingBatchStmt{}, nil
}

func (c *lockingBatchConn) Close() error {
	return nil
}

func (c *lockingBatchConn) Begin() (driver.Tx, error) {
	return &lockingBatchTx{state: c.state}, nil
}

func (c *lockingBatchConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return &lockingBatchTx{state: c.state}, nil
}

func (c *lockingBatchConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.mu.Lock()

	defer c.state.mu.Unlock()

	c.state.queries = append(c.state.queries, query)

	if strings.Contains(query, "SET pending_jobs = ? WHERE id = ?") {
		c.state.pending = namedInt(args[0])
	}

	if strings.Contains(query, "SET failed_jobs = ?, pending_jobs = ?, failed_job_ids = ?") {
		c.state.failed = namedInt(args[0])
		c.state.pending = namedInt(args[1])

		if v, ok := args[2].Value.(string); ok {
			_ = json.Unmarshal([]byte(v), &c.state.failedIDs)
		}
	}

	return driver.RowsAffected(1), nil
}

func (c *lockingBatchConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	c.state.mu.Lock()

	defer c.state.mu.Unlock()

	c.state.queries = append(c.state.queries, query)

	failedIDs, _ := json.Marshal(c.state.failedIDs)

	return &lockingBatchRows{
		columns: []string{"pending_jobs", "failed_jobs", "failed_job_ids"},
		values:  []driver.Value{int64(c.state.pending), int64(c.state.failed), string(failedIDs)},
	}, nil
}

func (tx *lockingBatchTx) Commit() error {
	tx.state.mu.Lock()

	defer tx.state.mu.Unlock()

	tx.state.commits++

	return nil
}

func (tx *lockingBatchTx) Rollback() error {
	tx.state.mu.Lock()

	defer tx.state.mu.Unlock()

	tx.state.rollbacks++

	return nil
}

func (r *lockingBatchRows) Columns() []string {
	return r.columns
}

func (r *lockingBatchRows) Close() error {
	return nil
}

func (r *lockingBatchRows) Next(dest []driver.Value) error {
	if r.read {
		return io.EOF
	}

	copy(dest, r.values)
	r.read = true

	return nil
}

func (lockingBatchStmt) Close() error {
	return nil
}

func (lockingBatchStmt) NumInput() int {
	return -1
}

func (lockingBatchStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}

func (lockingBatchStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return &lockingBatchRows{}, nil
}

func (s *lockingBatchDriverState) reset(pending, failed int, failedIDs []string) {
	s.mu.Lock()

	defer s.mu.Unlock()

	s.pending = pending
	s.failed = failed
	s.failedIDs = append([]string(nil), failedIDs...)
	s.queries = nil
	s.commits = 0
	s.rollbacks = 0
}

func namedInt(value driver.NamedValue) int {
	switch v := value.Value.(type) {
	case int64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func hasQuery(queries []string, needle string) bool {
	for _, query := range queries {
		if strings.Contains(query, needle) {
			return true
		}
	}

	return false
}

func TestDatabaseBatchRepositoryStore(t *testing.T) {
	db := newMockDBExecutor()
	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	batch := &bus.Batch{
		ID:          "batch-1",
		Name:        "test-batch",
		TotalJobs:   5,
		PendingJobs: 5,
		Options:     map[string]any{},
		CreatedAt:   time.Now(),
	}

	err := repo.Store(context.Background(), batch)

	if err != nil {
		t.Fatal(err)
	}

	db.mu.Lock()

	defer db.mu.Unlock()

	if len(db.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(db.execCalls))
	}

	call := db.execCalls[0]

	if call.Args[0] != "batch-1" {
		t.Errorf("expected batch ID 'batch-1', got %v", call.Args[0])
	}

	if call.Args[1] != "test-batch" {
		t.Errorf("expected batch name 'test-batch', got %v", call.Args[1])
	}
}

func TestDatabaseBatchRepositoryStoreSerializesJSON(t *testing.T) {
	db := newMockDBExecutor()
	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	batch := &bus.Batch{
		ID:           "batch-2",
		Name:         "json-test",
		FailedJobIDs: []string{"j1", "j2"},
		Options:      map[string]any{"allowFailures": true},
		CreatedAt:    time.Now(),
	}

	err := repo.Store(context.Background(), batch)

	if err != nil {
		t.Fatal(err)
	}

	db.mu.Lock()

	defer db.mu.Unlock()

	call := db.execCalls[0]

	failedJSON := call.Args[5].(string)

	var failedIDs []string

	if err := json.Unmarshal([]byte(failedJSON), &failedIDs); err != nil {
		t.Fatalf("failed to unmarshal failed_job_ids: %v", err)
	}

	if len(failedIDs) != 2 || failedIDs[0] != "j1" {
		t.Errorf("expected [j1, j2], got %v", failedIDs)
	}

	optsJSON := call.Args[6].(string)

	var opts map[string]any

	if err := json.Unmarshal([]byte(optsJSON), &opts); err != nil {
		t.Fatalf("failed to unmarshal options: %v", err)
	}

	if opts["allowFailures"] != true {
		t.Errorf("expected allowFailures true, got %v", opts["allowFailures"])
	}
}

func TestDatabaseBatchRepositoryStoreError(t *testing.T) {
	db := newMockDBExecutor()
	db.execErr = fmt.Errorf("db error")

	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	batch := &bus.Batch{ID: "batch-err", Options: map[string]any{}, CreatedAt: time.Now()}

	err := repo.Store(context.Background(), batch)

	if err == nil {
		t.Error("expected error from Store")
	}
}

func TestDatabaseBatchRepositoryIncrementTotalJobs(t *testing.T) {
	db := newMockDBExecutor()
	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	err := repo.IncrementTotalJobs(context.Background(), "batch-1", 3)

	if err != nil {
		t.Fatal(err)
	}

	db.mu.Lock()

	defer db.mu.Unlock()

	if len(db.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(db.execCalls))
	}

	if db.execCalls[0].Args[0] != 3 {
		t.Errorf("expected amount 3, got %v", db.execCalls[0].Args[0])
	}
}

func TestDatabaseBatchRepositoryDecrementPendingJobsLocksRow(t *testing.T) {
	sqlLockingState.reset(2, 0, nil)

	db, err := sql.Open("alloy-bus-locking-test", "")

	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	counts, err := repo.DecrementPendingJobs(context.Background(), "batch-1")

	if err != nil {
		t.Fatal(err)
	}

	if counts.PendingJobs != 1 || counts.FailedJobs != 0 {
		t.Fatalf("expected counts pending=1 failed=0, got pending=%d failed=%d", counts.PendingJobs, counts.FailedJobs)
	}

	sqlLockingState.mu.Lock()

	defer sqlLockingState.mu.Unlock()

	if sqlLockingState.pending != 1 {
		t.Fatalf("expected stored pending count 1, got %d", sqlLockingState.pending)
	}

	if sqlLockingState.commits != 1 || sqlLockingState.rollbacks != 0 {
		t.Fatalf("expected one commit and no rollbacks, got commits=%d rollbacks=%d", sqlLockingState.commits, sqlLockingState.rollbacks)
	}

	if !hasQuery(sqlLockingState.queries, "FOR UPDATE") {
		t.Fatalf("expected row lock query, got %v", sqlLockingState.queries)
	}
}

func TestDatabaseBatchRepositoryIncrementFailedJobsLocksRowAndAppendsID(t *testing.T) {
	sqlLockingState.reset(2, 1, []string{"failed-1"})

	db, err := sql.Open("alloy-bus-locking-test", "")

	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	counts, err := repo.IncrementFailedJobs(context.Background(), "batch-1", "failed-2")

	if err != nil {
		t.Fatal(err)
	}

	if counts.PendingJobs != 1 || counts.FailedJobs != 2 {
		t.Fatalf("expected counts pending=1 failed=2, got pending=%d failed=%d", counts.PendingJobs, counts.FailedJobs)
	}

	sqlLockingState.mu.Lock()

	defer sqlLockingState.mu.Unlock()

	if sqlLockingState.pending != 1 || sqlLockingState.failed != 2 {
		t.Fatalf("expected stored pending=1 failed=2, got pending=%d failed=%d", sqlLockingState.pending, sqlLockingState.failed)
	}

	if len(sqlLockingState.failedIDs) != 2 || sqlLockingState.failedIDs[0] != "failed-1" || sqlLockingState.failedIDs[1] != "failed-2" {
		t.Fatalf("expected appended failed IDs, got %v", sqlLockingState.failedIDs)
	}

	if sqlLockingState.commits != 1 || sqlLockingState.rollbacks != 0 {
		t.Fatalf("expected one commit and no rollbacks, got commits=%d rollbacks=%d", sqlLockingState.commits, sqlLockingState.rollbacks)
	}

	if !hasQuery(sqlLockingState.queries, "FOR UPDATE") {
		t.Fatalf("expected row lock query, got %v", sqlLockingState.queries)
	}
}

func TestDatabaseBatchRepositoryMarkAsFinished(t *testing.T) {
	db := newMockDBExecutor()
	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	err := repo.MarkAsFinished(context.Background(), "batch-1")

	if err != nil {
		t.Fatal(err)
	}

	db.mu.Lock()

	defer db.mu.Unlock()

	if len(db.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(db.execCalls))
	}

	if db.execCalls[0].Args[1] != "batch-1" {
		t.Errorf("expected batch ID 'batch-1', got %v", db.execCalls[0].Args[1])
	}
}

func TestDatabaseBatchRepositoryCancel(t *testing.T) {
	db := newMockDBExecutor()
	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	err := repo.Cancel(context.Background(), "batch-1")

	if err != nil {
		t.Fatal(err)
	}

	db.mu.Lock()

	defer db.mu.Unlock()

	if len(db.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(db.execCalls))
	}
}

func TestDatabaseBatchRepositoryDelete(t *testing.T) {
	db := newMockDBExecutor()
	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	err := repo.Delete(context.Background(), "batch-1")

	if err != nil {
		t.Fatal(err)
	}

	db.mu.Lock()

	defer db.mu.Unlock()

	if len(db.execCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(db.execCalls))
	}
}

func TestDatabaseBatchRepositoryPrune(t *testing.T) {
	db := newMockDBExecutor()

	callCount := 0
	db.execResult = &mockDynamicResult{fn: func() int64 {
		callCount++

		if callCount == 1 {
			return 1000
		}

		return 0
	}}

	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	total, err := repo.Prune(context.Background(), time.Now().Add(-24*time.Hour))

	if err != nil {
		t.Fatal(err)
	}

	if total != 1000 {
		t.Errorf("expected 1000 pruned, got %d", total)
	}
}

func TestDatabaseBatchRepositoryPruneCancelled(t *testing.T) {
	db := newMockDBExecutor()
	db.execResult = &mockResult{rowsAffected: 0}

	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	total, err := repo.PruneCancelled(context.Background(), time.Now())

	if err != nil {
		t.Fatal(err)
	}

	if total != 0 {
		t.Errorf("expected 0 pruned, got %d", total)
	}
}

func TestDatabaseBatchRepositoryPruneUnfinished(t *testing.T) {
	db := newMockDBExecutor()
	db.execResult = &mockResult{rowsAffected: 0}

	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	total, err := repo.PruneUnfinished(context.Background(), time.Now())

	if err != nil {
		t.Fatal(err)
	}

	if total != 0 {
		t.Errorf("expected 0 pruned, got %d", total)
	}
}

func TestDatabaseBatchRepositoryTransactionWithoutTransactor(t *testing.T) {
	db := newMockDBExecutor()
	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	called := false
	err := repo.Transaction(context.Background(), func(r bus.BatchRepository) error {
		called = true

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if !called {
		t.Error("expected fn to be called")
	}
}

func TestDatabaseBatchRepositoryDefaultTable(t *testing.T) {
	db := newMockDBExecutor()
	repo := bus.NewDatabaseBatchRepository(db, "")

	batch := &bus.Batch{ID: "b1", Options: map[string]any{}, CreatedAt: time.Now()}
	_ = repo.Store(context.Background(), batch)

	db.mu.Lock()

	defer db.mu.Unlock()

	if len(db.execCalls) == 0 {
		t.Fatal("expected at least 1 call")
	}

	query := db.execCalls[0].Query

	if len(query) < 20 {
		t.Error("query too short")
	}
}

func (r *mockDynamicResult) LastInsertId() (int64, error) { return 0, nil }
func (r *mockDynamicResult) RowsAffected() (int64, error) { return r.fn(), nil }

func TestDatabaseBatchRepositoryRollBack(t *testing.T) {
	db := newMockDBExecutor()
	repo := bus.NewDatabaseBatchRepository(db, "job_batches")

	err := repo.RollBack(context.Background())

	if err != nil {
		t.Errorf("expected no error from RollBack, got %v", err)
	}
}

// Ensure DatabaseBatchRepository satisfies the interfaces.
var _ bus.BatchRepository = (*bus.DatabaseBatchRepository)(nil)

// Use a driver.Value type assertion to suppress unused import warning.
var _ driver.Value = driver.Value(nil)

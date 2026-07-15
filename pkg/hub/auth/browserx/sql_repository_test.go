package browserx

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type sessionSQLRecord struct {
	id           string
	userID       string
	ipAddress    string
	userAgent    string
	lastActivity int64
}

type sessionSQLDB struct {
	lastQuery string
	queryArgs []any
	lastExec  string
	execArgs  []any
	records   []sessionSQLRecord
	queryErr  error
	scanErr   error
	rowsErr   error
	execErr   error
	// lastRows tracks the rows handle returned by Query so tests can assert
	// the repository always closes it (connection-leak guard).
	lastRows *sessionSQLRows
}

type sessionSQLRows struct {
	records []sessionSQLRecord
	pos     int
	scanErr error
	rowsErr error
	closed  bool
}

func TestSQLRepositoryFindForUserMapsRowsAndBindsUser(t *testing.T) {
	lastActive := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC).Unix()
	db := &sessionSQLDB{
		records: []sessionSQLRecord{{
			id:           "sess-1",
			userID:       "42",
			ipAddress:    "203.0.113.7",
			userAgent:    "Firefox",
			lastActivity: lastActive,
		}},
	}
	repo := NewSQLRepository(db, "sessions")

	sessions, err := repo.FindForUser(context.Background(), "42")

	if err != nil {
		t.Fatal(err)
	}

	want := "SELECT id, user_id, ip_address, user_agent, last_activity FROM sessions WHERE user_id = $1 ORDER BY last_activity DESC"

	if !strings.Contains(db.lastQuery, want) {
		t.Fatalf("query = %q", db.lastQuery)
	}

	if len(db.queryArgs) != 1 || db.queryArgs[0] != "42" {
		t.Fatalf("query args = %#v", db.queryArgs)
	}

	if len(sessions) != 1 {
		t.Fatalf("sessions = %#v", sessions)
	}

	got := sessions[0]

	if got.ID != "sess-1" || got.UserID != "42" || got.IPAddress != "203.0.113.7" || got.UserAgent != "Firefox" {
		t.Fatalf("session = %#v", got)
	}

	if !got.LastActiveAt.Equal(time.Unix(lastActive, 0)) {
		t.Fatalf("last active = %v", got.LastActiveAt)
	}

	if db.lastRows == nil || !db.lastRows.closed {
		t.Fatal("expected rows to be closed")
	}
}

func TestSQLRepositoryFindForUserReturnsEmptySliceWhenNoRows(t *testing.T) {
	db := &sessionSQLDB{}
	repo := NewSQLRepository(db, "sessions")

	sessions, err := repo.FindForUser(context.Background(), "42")

	if err != nil {
		t.Fatal(err)
	}

	if sessions == nil || len(sessions) != 0 {
		t.Fatalf("sessions = %#v", sessions)
	}

	if db.lastRows == nil || !db.lastRows.closed {
		t.Fatal("expected rows to be closed even with no rows returned")
	}
}

func TestSQLRepositoryFindForUserPropagatesErrors(t *testing.T) {
	record := sessionSQLRecord{id: "sess-1", userID: "42"}

	cases := []struct {
		name string
		db   *sessionSQLDB
		want error
	}{
		{name: "query error", db: &sessionSQLDB{queryErr: errors.New("query failed")}, want: errors.New("query failed")},
		{name: "scan error", db: &sessionSQLDB{records: []sessionSQLRecord{record}, scanErr: errors.New("scan failed")}, want: errors.New("scan failed")},
		{name: "rows error", db: &sessionSQLDB{records: []sessionSQLRecord{record}, rowsErr: errors.New("rows failed")}, want: errors.New("rows failed")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewSQLRepository(tc.db, "sessions")

			sessions, err := repo.FindForUser(context.Background(), "42")

			if err == nil || err.Error() != tc.want.Error() {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}

			if sessions != nil {
				t.Fatalf("sessions = %#v", sessions)
			}

			// On the query-error path no rows handle is ever produced; every
			// other error path must still close the handle it received.
			if tc.name != "query error" {
				if tc.db.lastRows == nil || !tc.db.lastRows.closed {
					t.Fatal("expected rows to be closed on error paths")
				}
			}
		})
	}
}

func TestSQLRepositoryRevokeDeletesUserScopedSession(t *testing.T) {
	db := &sessionSQLDB{}
	repo := NewSQLRepository(db, "sessions")

	if err := repo.Revoke(context.Background(), "42", "sess-1"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.lastExec, "DELETE FROM sessions WHERE user_id = $1 AND id = $2") {
		t.Fatalf("exec = %q", db.lastExec)
	}

	if len(db.execArgs) != 2 || db.execArgs[0] != "42" || db.execArgs[1] != "sess-1" {
		t.Fatalf("exec args = %#v", db.execArgs)
	}
}

func TestSQLRepositoryRevokeOtherExcludesCurrentSession(t *testing.T) {
	db := &sessionSQLDB{}
	repo := NewSQLRepository(db, "sessions")

	if err := repo.RevokeOther(context.Background(), "42", "sess-current"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.lastExec, "DELETE FROM sessions WHERE user_id = $1 AND id <> $2") {
		t.Fatalf("exec = %q", db.lastExec)
	}

	if len(db.execArgs) != 2 || db.execArgs[0] != "42" || db.execArgs[1] != "sess-current" {
		t.Fatalf("exec args = %#v", db.execArgs)
	}
}

func TestSQLRepositoryMutationsPropagateExecErrors(t *testing.T) {
	db := &sessionSQLDB{execErr: errors.New("exec failed")}
	repo := NewSQLRepository(db, "sessions")

	if err := repo.Revoke(context.Background(), "42", "sess-1"); err == nil || err.Error() != "exec failed" {
		t.Fatalf("revoke err = %v", err)
	}

	if err := repo.RevokeOther(context.Background(), "42", "sess-1"); err == nil || err.Error() != "exec failed" {
		t.Fatalf("revoke other err = %v", err)
	}
}

func TestSQLRepositorySanitizesUnsafeTableNames(t *testing.T) {
	cases := []struct {
		name  string
		table string
		want  string
	}{
		{name: "sql injection", table: "sessions; DROP TABLE users", want: "sessions"},
		{name: "empty", table: "", want: "sessions"},
		{name: "safe dotted", table: "auth.sessions", want: "auth.sessions"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := &sessionSQLDB{}
			repo := NewSQLRepository(db, tc.table)

			if _, err := repo.FindForUser(context.Background(), "42"); err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(db.lastQuery, " FROM "+tc.want+" WHERE") {
				t.Fatalf("query = %q", db.lastQuery)
			}

			if strings.Contains(db.lastQuery, "DROP TABLE") {
				t.Fatalf("query leaked unsafe table name: %q", db.lastQuery)
			}
		})
	}
}

func (db *sessionSQLDB) Query(_ context.Context, query string, args ...any) (SQLRows, error) {
	db.lastQuery = query
	db.queryArgs = args

	if db.queryErr != nil {
		return nil, db.queryErr
	}

	rows := &sessionSQLRows{records: db.records, scanErr: db.scanErr, rowsErr: db.rowsErr}
	db.lastRows = rows

	return rows, nil
}

func (db *sessionSQLDB) Exec(_ context.Context, query string, args ...any) error {
	db.lastExec = query
	db.execArgs = args

	return db.execErr
}

func (r *sessionSQLRows) Next() bool {
	return r.pos < len(r.records)
}

func (r *sessionSQLRows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}

	record := r.records[r.pos]
	r.pos++

	*dest[0].(*string) = record.id
	*dest[1].(*string) = record.userID
	*dest[2].(*string) = record.ipAddress
	*dest[3].(*string) = record.userAgent
	*dest[4].(*int64) = record.lastActivity

	return nil
}

func (r *sessionSQLRows) Close() error {
	r.closed = true

	return nil
}

func (r *sessionSQLRows) Err() error {
	return r.rowsErr
}

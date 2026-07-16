package passwords

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type resetSQLDB struct {
	lastQuery string
	queryArgs []any
	lastExec  string
	execArgs  []any
	row       resetSQLRow
	execErr   error
}

type resetSQLRow struct {
	token     string
	createdAt time.Time
	err       error
}

func TestSQLRepositoryCreateStoresHashedTokenNotPlaintext(t *testing.T) {
	db := &resetSQLDB{}
	repo := NewSQLRepository(db, "password_reset_tokens", time.Hour)

	token, err := repo.Create(context.Background(), "user@example.com")

	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Fatal("expected plaintext token to be returned")
	}

	if !strings.HasPrefix(db.lastExec, "INSERT INTO password_reset_tokens (email, token, created_at)") {
		t.Fatalf("exec = %q", db.lastExec)
	}

	if !strings.Contains(db.lastExec, "ON CONFLICT (email) DO UPDATE SET token=$2, created_at=$3") {
		t.Fatalf("exec = %q", db.lastExec)
	}

	if len(db.execArgs) != 3 || db.execArgs[0] != "user@example.com" {
		t.Fatalf("exec args = %#v", db.execArgs)
	}

	stored, ok := db.execArgs[1].(string)

	if !ok || stored == token {
		t.Fatalf("SQL repository stored plaintext token: %#v", db.execArgs[1])
	}

	if stored != hashToken(token) {
		t.Fatalf("stored token = %q, want hash of plaintext", stored)
	}

	createdAt, ok := db.execArgs[2].(time.Time)

	if !ok || time.Since(createdAt) > time.Minute {
		t.Fatalf("created at = %#v", db.execArgs[2])
	}
}

func TestSQLRepositoryCreatePropagatesExecError(t *testing.T) {
	db := &resetSQLDB{execErr: errors.New("exec failed")}
	repo := NewSQLRepository(db, "password_reset_tokens", time.Hour)

	token, err := repo.Create(context.Background(), "user@example.com")

	if err == nil || err.Error() != "exec failed" {
		t.Fatalf("err = %v", err)
	}

	if token != "" {
		t.Fatalf("token = %q", token)
	}
}

func TestSQLRepositoryExistsComparesHashesWithinExpiry(t *testing.T) {
	cases := []struct {
		name      string
		stored    string
		createdAt time.Time
		token     string
		scanErr   error
		want      bool
	}{
		{name: "matching hash within expiry", stored: hashToken("secret"), createdAt: time.Now(), token: "secret", want: true},
		{name: "wrong token", stored: hashToken("secret"), createdAt: time.Now(), token: "other", want: false},
		{name: "plaintext stored never matches", stored: "secret", createdAt: time.Now(), token: "secret", want: false},
		{name: "expired token", stored: hashToken("secret"), createdAt: time.Now().Add(-2 * time.Hour), token: "secret", want: false},
		{name: "scan error", scanErr: errors.New("no rows"), token: "secret", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := &resetSQLDB{row: resetSQLRow{token: tc.stored, createdAt: tc.createdAt, err: tc.scanErr}}
			repo := NewSQLRepository(db, "password_reset_tokens", time.Hour)

			if got := repo.Exists(context.Background(), "user@example.com", tc.token); got != tc.want {
				t.Fatalf("exists = %v, want %v", got, tc.want)
			}

			if !strings.Contains(db.lastQuery, "SELECT token, created_at FROM password_reset_tokens WHERE email = $1 LIMIT 1") {
				t.Fatalf("query = %q", db.lastQuery)
			}

			if len(db.queryArgs) != 1 || db.queryArgs[0] != "user@example.com" {
				t.Fatalf("query args = %#v", db.queryArgs)
			}
		})
	}
}

func TestSQLRepositoryRecentlyCreatedUsesTimeWindow(t *testing.T) {
	cases := []struct {
		name      string
		createdAt time.Time
		within    time.Duration
		scanErr   error
		want      bool
	}{
		{name: "created inside window", createdAt: time.Now().Add(-time.Minute), within: time.Hour, want: true},
		{name: "created outside window", createdAt: time.Now().Add(-2 * time.Hour), within: time.Hour, want: false},
		{name: "scan error", scanErr: errors.New("no rows"), within: time.Hour, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := &resetSQLDB{row: resetSQLRow{createdAt: tc.createdAt, err: tc.scanErr}}
			repo := NewSQLRepository(db, "password_reset_tokens", time.Hour)

			if got := repo.RecentlyCreated(context.Background(), "user@example.com", tc.within); got != tc.want {
				t.Fatalf("recently created = %v, want %v", got, tc.want)
			}

			if !strings.Contains(db.lastQuery, "SELECT created_at FROM password_reset_tokens WHERE email = $1 LIMIT 1") {
				t.Fatalf("query = %q", db.lastQuery)
			}

			if len(db.queryArgs) != 1 || db.queryArgs[0] != "user@example.com" {
				t.Fatalf("query args = %#v", db.queryArgs)
			}
		})
	}
}

func TestSQLRepositoryDeleteRemovesTokensByEmail(t *testing.T) {
	db := &resetSQLDB{}
	repo := NewSQLRepository(db, "password_reset_tokens", time.Hour)

	if err := repo.Delete(context.Background(), "user@example.com"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.lastExec, "DELETE FROM password_reset_tokens WHERE email = $1") {
		t.Fatalf("exec = %q", db.lastExec)
	}

	if len(db.execArgs) != 1 || db.execArgs[0] != "user@example.com" {
		t.Fatalf("exec args = %#v", db.execArgs)
	}

	db.execErr = errors.New("exec failed")

	if err := repo.Delete(context.Background(), "user@example.com"); err == nil || err.Error() != "exec failed" {
		t.Fatalf("err = %v", err)
	}
}

func TestSQLRepositoryDeleteExpiredBindsExpiryCutoff(t *testing.T) {
	expiry := time.Hour
	db := &resetSQLDB{}
	repo := NewSQLRepository(db, "password_reset_tokens", expiry)

	before := time.Now().UTC().Add(-expiry)

	if err := repo.DeleteExpired(context.Background()); err != nil {
		t.Fatal(err)
	}

	after := time.Now().UTC().Add(-expiry)

	if !strings.Contains(db.lastExec, "DELETE FROM password_reset_tokens WHERE created_at < $1") {
		t.Fatalf("exec = %q", db.lastExec)
	}

	if len(db.execArgs) != 1 {
		t.Fatalf("exec args = %#v", db.execArgs)
	}

	cutoff, ok := db.execArgs[0].(time.Time)

	if !ok || cutoff.Before(before) || cutoff.After(after) {
		t.Fatalf("cutoff = %#v, want between %v and %v", db.execArgs[0], before, after)
	}

	db.execErr = errors.New("exec failed")

	if err := repo.DeleteExpired(context.Background()); err == nil || err.Error() != "exec failed" {
		t.Fatalf("err = %v", err)
	}
}

func TestSQLRepositorySanitizesUnsafeTableNames(t *testing.T) {
	cases := []struct {
		name  string
		table string
		want  string
	}{
		{name: "sql injection", table: "password_resets; DROP TABLE users", want: "password_reset_tokens"},
		{name: "empty", table: "", want: "password_reset_tokens"},
		{name: "safe dotted", table: "auth.password_reset_tokens", want: "auth.password_reset_tokens"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := &resetSQLDB{}
			repo := NewSQLRepository(db, tc.table, time.Hour)

			if err := repo.Delete(context.Background(), "user@example.com"); err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(db.lastExec, "DELETE FROM "+tc.want+" WHERE email = $1") {
				t.Fatalf("exec = %q", db.lastExec)
			}

			if strings.Contains(db.lastExec, "DROP TABLE") {
				t.Fatalf("exec leaked unsafe table name: %q", db.lastExec)
			}
		})
	}
}

func (db *resetSQLDB) QueryRow(_ context.Context, query string, args ...any) SQLRow {
	db.lastQuery = query
	db.queryArgs = args

	return db.row
}

func (db *resetSQLDB) Exec(_ context.Context, query string, args ...any) error {
	db.lastExec = query
	db.execArgs = args

	return db.execErr
}

func (r resetSQLRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	if len(dest) == 2 {
		*dest[0].(*string) = r.token
		*dest[1].(*time.Time) = r.createdAt

		return nil
	}

	*dest[0].(*time.Time) = r.createdAt

	return nil
}

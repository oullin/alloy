package browserx_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"hara.sh/alloy/auth/browserx"
)

type sessionSQLDB struct {
	lastQuery string
	lastExec  string
	rows      []browserx.Session
}

type sessionSQLRows struct {
	rows []browserx.Session
	pos  int
}

func TestServiceListsSessionsAndMarksCurrent(t *testing.T) {
	repo := browserx.NewMemoryRepository(
		browserx.Session{ID: "current", UserID: "1"},
		browserx.Session{ID: "other", UserID: "1"},
		browserx.Session{ID: "other-user", UserID: "2"},
	)
	service := browserx.NewService(repo)

	sessions, err := service.List(context.Background(), "1", "current")

	if err != nil {
		t.Fatal(err)
	}

	if len(sessions) != 2 {
		t.Fatalf("sessions = %d, want 2", len(sessions))
	}

	for _, session := range sessions {
		if session.ID == "current" && !session.Current {
			t.Fatal("expected current session to be marked")
		}

		if session.ID == "other" && session.Current {
			t.Fatal("expected other session not to be current")
		}
	}
}

func TestMemoryRepositoryRevokesSessionAndOtherSessions(t *testing.T) {
	repo := browserx.NewMemoryRepository(
		browserx.Session{ID: "current", UserID: "1"},
		browserx.Session{ID: "other", UserID: "1"},
		browserx.Session{ID: "other-user", UserID: "2"},
	)

	if err := repo.Revoke(context.Background(), "1", "other"); err != nil {
		t.Fatal(err)
	}

	sessions, err := repo.FindForUser(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if len(sessions) != 1 || sessions[0].ID != "current" {
		t.Fatalf("sessions = %#v", sessions)
	}

	if err := repo.RevokeOther(context.Background(), "1", "current"); err != nil {
		t.Fatal(err)
	}

	sessions, err = repo.FindForUser(context.Background(), "2")

	if err != nil {
		t.Fatal(err)
	}

	if len(sessions) != 1 || sessions[0].ID != "other-user" {
		t.Fatalf("other user sessions = %#v", sessions)
	}
}

func TestSQLRepositoryUsesSafeTableAndUserScopedDeletes(t *testing.T) {
	db := &sessionSQLDB{
		rows: []browserx.Session{{
			ID:           "current",
			UserID:       "1",
			IPAddress:    "127.0.0.1",
			UserAgent:    "Agent",
			LastActiveAt: time.Unix(100, 0),
		}},
	}
	repo := browserx.NewSQLRepository(db, "sessions; DROP TABLE users")

	sessions, err := repo.FindForUser(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if len(sessions) != 1 || sessions[0].ID != "current" {
		t.Fatalf("sessions = %#v", sessions)
	}

	if !strings.Contains(db.lastQuery, "FROM sessions ") {
		t.Fatalf("query = %q", db.lastQuery)
	}

	if err := repo.Revoke(context.Background(), "1", "current"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.lastExec, "WHERE user_id = $1 AND id = $2") {
		t.Fatalf("revoke query = %q", db.lastExec)
	}

	if err := repo.RevokeOther(context.Background(), "1", "current"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.lastExec, "WHERE user_id = $1 AND id <> $2") {
		t.Fatalf("revoke other query = %q", db.lastExec)
	}
}

func (db *sessionSQLDB) Query(_ context.Context, query string, _ ...any) (browserx.SQLRows, error) {
	db.lastQuery = query

	return &sessionSQLRows{rows: db.rows}, nil
}

func (db *sessionSQLDB) Exec(_ context.Context, query string, _ ...any) error {
	db.lastExec = query

	return nil
}

func (r *sessionSQLRows) Next() bool {
	return r.pos < len(r.rows)
}

func (r *sessionSQLRows) Scan(dest ...any) error {
	session := r.rows[r.pos]
	r.pos++

	*dest[0].(*string) = session.ID
	*dest[1].(*string) = session.UserID
	*dest[2].(*string) = session.IPAddress
	*dest[3].(*string) = session.UserAgent
	*dest[4].(*int64) = session.LastActiveAt.Unix()

	return nil
}

func (r *sessionSQLRows) Close() error { return nil }
func (r *sessionSQLRows) Err() error   { return nil }

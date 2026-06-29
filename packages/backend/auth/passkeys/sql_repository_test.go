package passkeys

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

type passkeySQLDB struct {
	lastQuery   string
	lastExec    string
	execArgs    []any
	handle      []byte
	userID      string
	credentials []webauthn.Credential
	session     webauthn.SessionData
}

type passkeySQLRow struct {
	handle    []byte
	userID    string
	payload   []byte
	expiresAt time.Time
	err       error
}

type passkeySQLRows struct {
	credentials []webauthn.Credential
	pos         int
}

func TestSQLRepositoryCreatesHandleWithSafeTableFallback(t *testing.T) {
	db := &passkeySQLDB{}
	repo := NewSQLRepository(db, "example.com", "webauthn_users; DROP TABLE users", "webauthn_credentials; DROP TABLE users")

	handle, err := repo.GetOrCreateUserHandle(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if len(handle) == 0 {
		t.Fatal("expected generated handle")
	}

	if !strings.Contains(db.lastQuery, "FROM webauthn_users ") {
		t.Fatalf("query = %q", db.lastQuery)
	}

	if !strings.HasPrefix(db.lastExec, "INSERT INTO webauthn_users ") {
		t.Fatalf("exec = %q", db.lastExec)
	}

	if db.execArgs[0] != "example.com" || db.execArgs[1] != "1" {
		t.Fatalf("exec args = %#v", db.execArgs)
	}
}

func TestSQLRepositoryFindsExistingHandleAndUserIDByHandle(t *testing.T) {
	handle := []byte("handle")
	db := &passkeySQLDB{handle: handle, userID: "1"}
	repo := NewSQLRepository(db, "example.com", "auth.webauthn_users", "auth.webauthn_credentials")

	got, err := repo.GetOrCreateUserHandle(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "handle" {
		t.Fatalf("handle = %q", got)
	}

	userID, err := repo.UserIDByHandle(context.Background(), handle)

	if err != nil {
		t.Fatal(err)
	}

	if userID != "1" {
		t.Fatalf("userID = %q", userID)
	}

	if !strings.Contains(db.lastQuery, "FROM auth.webauthn_users") {
		t.Fatalf("query = %q", db.lastQuery)
	}
}

func TestSQLRepositorySavesUpdatesAndReadsCredentials(t *testing.T) {
	credential := webauthn.Credential{ID: []byte("credential"), PublicKey: []byte("public-key")}
	db := &passkeySQLDB{}
	repo := NewSQLRepository(db, "example.com", "webauthn_users", "webauthn_credentials")

	if err := repo.SaveCredential(context.Background(), "1", credential); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(db.lastExec, "INSERT INTO webauthn_credentials ") {
		t.Fatalf("insert = %q", db.lastExec)
	}

	if string(db.execArgs[2].([]byte)) != "credential" {
		t.Fatalf("credential id arg = %#v", db.execArgs[2])
	}

	db.credentials = []webauthn.Credential{credential}
	found, err := repo.CredentialsByUser(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if len(found) != 1 || string(found[0].ID) != "credential" {
		t.Fatalf("found = %#v", found)
	}

	credential.Authenticator.SignCount = 10

	if err := repo.UpdateCredential(context.Background(), "1", credential); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(db.lastExec, "UPDATE webauthn_credentials ") {
		t.Fatalf("update = %q", db.lastExec)
	}
}

func TestSQLSessionStoreStoresLoadsAndDeletesSessions(t *testing.T) {
	expires := time.Now().Add(time.Hour)
	db := &passkeySQLDB{}
	store := NewSQLSessionStore(db, "example.com", "webauthn_sessions; DROP")

	session := webauthn.SessionData{Challenge: "challenge", Expires: expires}

	if err := store.Put(context.Background(), "session", session); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(db.lastExec, "INSERT INTO webauthn_sessions ") {
		t.Fatalf("insert = %q", db.lastExec)
	}

	db.session = session
	got, err := store.Get(context.Background(), "session")

	if err != nil {
		t.Fatal(err)
	}

	if got.Challenge != "challenge" {
		t.Fatalf("challenge = %q", got.Challenge)
	}

	if err := store.Delete(context.Background(), "session"); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(db.lastExec, "DELETE FROM webauthn_sessions ") {
		t.Fatalf("delete = %q", db.lastExec)
	}
}

func TestSQLSessionStoreRejectsExpiredSessionAndDeletesIt(t *testing.T) {
	db := &passkeySQLDB{session: webauthn.SessionData{Challenge: "expired", Expires: time.Now().Add(-time.Hour)}}
	store := NewSQLSessionStore(db, "example.com", "webauthn_sessions")

	if _, err := store.Get(context.Background(), "session"); err != ErrCredentialNotFound {
		t.Fatalf("err = %v, want ErrCredentialNotFound", err)
	}

	if !strings.HasPrefix(db.lastExec, "DELETE FROM webauthn_sessions ") {
		t.Fatalf("expected expired session delete, got %q", db.lastExec)
	}
}

func (db *passkeySQLDB) QueryRow(_ context.Context, query string, args ...any) SQLRow {
	db.lastQuery = query

	switch {
	case strings.Contains(query, "SELECT handle"):
		if db.handle == nil {
			return passkeySQLRow{err: ErrUserHandleNotFound}
		}

		return passkeySQLRow{handle: db.handle}
	case strings.Contains(query, "SELECT user_id"):
		if db.userID == "" {
			return passkeySQLRow{err: ErrUserHandleNotFound}
		}

		return passkeySQLRow{userID: db.userID}
	case strings.Contains(query, "SELECT payload"):
		if db.session.Challenge == "" {
			return passkeySQLRow{err: ErrCredentialNotFound}
		}

		payload, _ := encodeSession(db.session)

		return passkeySQLRow{payload: payload, expiresAt: db.session.Expires}
	default:
		return passkeySQLRow{err: ErrCredentialNotFound}
	}
}

func (db *passkeySQLDB) Query(_ context.Context, query string, _ ...any) (SQLRows, error) {
	db.lastQuery = query

	return &passkeySQLRows{credentials: db.credentials}, nil
}

func (db *passkeySQLDB) Exec(_ context.Context, query string, args ...any) error {
	db.lastExec = query
	db.execArgs = append([]any(nil), args...)

	return nil
}

func (r passkeySQLRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	switch {
	case r.handle != nil:
		*dest[0].(*[]byte) = append([]byte(nil), r.handle...)
	case r.userID != "":
		*dest[0].(*string) = r.userID
	default:
		*dest[0].(*[]byte) = append([]byte(nil), r.payload...)
		*dest[1].(*time.Time) = r.expiresAt
	}

	return nil
}

func (r *passkeySQLRows) Next() bool {
	return r.pos < len(r.credentials)
}

func (r *passkeySQLRows) Scan(dest ...any) error {
	payload, err := encodeCredential(r.credentials[r.pos])

	if err != nil {
		return err
	}

	r.pos++
	*dest[0].(*[]byte) = payload

	return nil
}

func (r *passkeySQLRows) Close() error { return nil }
func (r *passkeySQLRows) Err() error   { return nil }

func encodeSession(session webauthn.SessionData) ([]byte, error) {
	return json.Marshal(session)
}

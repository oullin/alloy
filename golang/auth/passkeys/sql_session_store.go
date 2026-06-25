package passkeys

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// SQLSessionStore stores WebAuthn ceremony session data server-side.
type SQLSessionStore struct {
	db    SQLQuerier
	rpID  string
	table string
	now   func() time.Time
}

// NewSQLSessionStore creates a SQLSessionStore.
func NewSQLSessionStore(db SQLQuerier, rpID, table string) *SQLSessionStore {
	if !isSafeSQLIdentifier(table) {
		table = "webauthn_sessions"
	}

	return &SQLSessionStore{db: db, rpID: rpID, table: table, now: time.Now}
}

func (s *SQLSessionStore) Put(ctx context.Context, key string, data webauthn.SessionData) error {
	payload, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return s.db.Exec(ctx,
		"INSERT INTO "+s.table+" (rpid, id, challenge, payload, expires_at) VALUES ($1, $2, $3, $4, $5) "+
			"ON CONFLICT (rpid, id) DO UPDATE SET challenge=$3, payload=$4, expires_at=$5",
		s.rpID, key, data.Challenge, payload, data.Expires,
	)
}

func (s *SQLSessionStore) Get(ctx context.Context, key string) (webauthn.SessionData, error) {
	row := s.db.QueryRow(ctx, "SELECT payload, expires_at FROM "+s.table+" WHERE rpid = $1 AND id = $2 LIMIT 1", s.rpID, key)

	var payload []byte

	var expiresAt time.Time

	if err := row.Scan(&payload, &expiresAt); err != nil {
		return webauthn.SessionData{}, ErrCredentialNotFound
	}

	if !expiresAt.After(s.now()) {
		_ = s.Delete(ctx, key)

		return webauthn.SessionData{}, ErrCredentialNotFound
	}

	var data webauthn.SessionData

	if err := json.Unmarshal(payload, &data); err != nil {
		return webauthn.SessionData{}, err
	}

	return data, nil
}

func (s *SQLSessionStore) Delete(ctx context.Context, key string) error {
	return s.db.Exec(ctx, "DELETE FROM "+s.table+" WHERE rpid = $1 AND id = $2", s.rpID, key)
}

// DeleteExpired removes expired WebAuthn ceremony sessions.
func (s *SQLSessionStore) DeleteExpired(ctx context.Context) error {
	return s.db.Exec(ctx, "DELETE FROM "+s.table+" WHERE rpid = $1 AND expires_at <= $2", s.rpID, s.now())
}

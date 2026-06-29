package passkeys

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-webauthn/webauthn/webauthn"
)

// SQLQuerier is the minimal SQL surface needed by passkey storage.
type SQLQuerier interface {
	QueryRow(ctx context.Context, query string, args ...any) SQLRow
	Query(ctx context.Context, query string, args ...any) (SQLRows, error)
	Exec(ctx context.Context, query string, args ...any) error
}

// SQLRow is a single SQL row.
type SQLRow interface {
	Scan(dest ...any) error
}

// SQLRows is a SQL result set.
type SQLRows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

// SQLRepository stores passkey user handles and credentials in SQL.
type SQLRepository struct {
	db               SQLQuerier
	rpID             string
	usersTable       string
	credentialsTable string
}

// NewSQLRepository creates a SQL-backed passkey repository.
func NewSQLRepository(db SQLQuerier, rpID, usersTable, credentialsTable string) *SQLRepository {
	if !isSafeSQLIdentifier(usersTable) {
		usersTable = "webauthn_users"
	}

	if !isSafeSQLIdentifier(credentialsTable) {
		credentialsTable = "webauthn_credentials"
	}

	return &SQLRepository{
		db:               db,
		rpID:             rpID,
		usersTable:       usersTable,
		credentialsTable: credentialsTable,
	}
}

func (r *SQLRepository) GetOrCreateUserHandle(ctx context.Context, userID string) ([]byte, error) {
	row := r.db.QueryRow(ctx, "SELECT handle FROM "+r.usersTable+" WHERE rpid = $1 AND user_id = $2 LIMIT 1", r.rpID, userID)

	var handle []byte

	if err := row.Scan(&handle); err == nil {
		return append([]byte(nil), handle...), nil
	}

	handle, err := generateUserHandle()

	if err != nil {
		return nil, err
	}

	if err := r.db.Exec(ctx, "INSERT INTO "+r.usersTable+" (rpid, user_id, handle) VALUES ($1, $2, $3)", r.rpID, userID, handle); err != nil {
		return nil, err
	}

	return append([]byte(nil), handle...), nil
}

func (r *SQLRepository) UserIDByHandle(ctx context.Context, handle []byte) (string, error) {
	row := r.db.QueryRow(ctx, "SELECT user_id FROM "+r.usersTable+" WHERE rpid = $1 AND handle = $2 LIMIT 1", r.rpID, handle)

	var userID string

	if err := row.Scan(&userID); err != nil {
		return "", ErrUserHandleNotFound
	}

	return userID, nil
}

func (r *SQLRepository) CredentialsByUser(ctx context.Context, userID string) ([]webauthn.Credential, error) {
	rows, err := r.db.Query(ctx, "SELECT credential FROM "+r.credentialsTable+" WHERE rpid = $1 AND user_id = $2", r.rpID, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	credentials := make([]webauthn.Credential, 0)

	for rows.Next() {
		var payload []byte

		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}

		credential, err := decodeCredential(payload)

		if err != nil {
			return nil, err
		}

		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return credentials, nil
}

func (r *SQLRepository) SaveCredential(ctx context.Context, userID string, credential webauthn.Credential) error {
	payload, err := encodeCredential(credential)

	if err != nil {
		return err
	}

	return r.db.Exec(ctx,
		"INSERT INTO "+r.credentialsTable+" (rpid, user_id, credential_id, credential) VALUES ($1, $2, $3, $4)",
		r.rpID, userID, credential.ID, payload,
	)
}

func (r *SQLRepository) UpdateCredential(ctx context.Context, userID string, credential webauthn.Credential) error {
	payload, err := encodeCredential(credential)

	if err != nil {
		return err
	}

	return r.db.Exec(ctx,
		"UPDATE "+r.credentialsTable+" SET credential = $1, last_used_at = CURRENT_TIMESTAMP WHERE rpid = $2 AND user_id = $3 AND credential_id = $4",
		payload, r.rpID, userID, credential.ID,
	)
}

func encodeCredential(credential webauthn.Credential) ([]byte, error) {
	return json.Marshal(credential)
}

func decodeCredential(payload []byte) (webauthn.Credential, error) {
	var credential webauthn.Credential

	if err := json.Unmarshal(payload, &credential); err != nil {
		return webauthn.Credential{}, err
	}

	return credential, nil
}

func isSafeSQLIdentifier(identifier string) bool {
	if identifier == "" {
		return false
	}

	parts := strings.Split(identifier, ".")

	for _, part := range parts {
		if part == "" {
			return false
		}

		for i, r := range part {
			if i == 0 {
				if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && r != '_' {
					return false
				}

				continue
			}

			if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
				return false
			}
		}
	}

	return true
}

package browserx

import (
	"context"
	"strings"
	"time"
)

// SQLQuerier is the minimal SQL surface for browser sessions.
type SQLQuerier interface {
	Query(ctx context.Context, query string, args ...any) (SQLRows, error)
	Exec(ctx context.Context, query string, args ...any) error
}

// SQLRows is a SQL result set.
type SQLRows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

// SQLRepository lists and revokes sessions from the session database table.
type SQLRepository struct {
	db    SQLQuerier
	table string
}

// NewSQLRepository creates a SQLRepository.
func NewSQLRepository(db SQLQuerier, table string) *SQLRepository {
	if !isSafeSQLIdentifier(table) {
		table = "sessions"
	}

	return &SQLRepository{db: db, table: table}
}

func (r *SQLRepository) FindForUser(ctx context.Context, userID string) ([]Session, error) {
	rows, err := r.db.Query(ctx,
		"SELECT id, user_id, ip_address, user_agent, last_activity FROM "+r.table+" WHERE user_id = $1 ORDER BY last_activity DESC",
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sessions := make([]Session, 0)

	for rows.Next() {
		var session Session

		var lastActivity int64

		if err := rows.Scan(&session.ID, &session.UserID, &session.IPAddress, &session.UserAgent, &lastActivity); err != nil {
			return nil, err
		}

		session.LastActiveAt = time.Unix(lastActivity, 0)
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SQLRepository) Revoke(ctx context.Context, userID, sessionID string) error {
	return r.db.Exec(ctx, "DELETE FROM "+r.table+" WHERE user_id = $1 AND id = $2", userID, sessionID)
}

func (r *SQLRepository) RevokeOther(ctx context.Context, userID, currentSessionID string) error {
	return r.db.Exec(ctx, "DELETE FROM "+r.table+" WHERE user_id = $1 AND id <> $2", userID, currentSessionID)
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

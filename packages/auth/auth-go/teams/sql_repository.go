package teams

import (
	"context"
	"strings"
	"time"
)

// SQLQuerier is the minimal SQL surface for team storage.
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

// SQLRepository stores teams in SQL.
type SQLRepository struct {
	db               SQLQuerier
	teamsTable       string
	membersTable     string
	currentTeamTable string
	now              func() time.Time
}

// NewSQLRepository creates a SQL-backed team repository.
func NewSQLRepository(db SQLQuerier, teamsTable, membersTable, currentTeamTable string) *SQLRepository {
	if !isSafeSQLIdentifier(teamsTable) {
		teamsTable = "teams"
	}

	if !isSafeSQLIdentifier(membersTable) {
		membersTable = "team_members"
	}

	if !isSafeSQLIdentifier(currentTeamTable) {
		currentTeamTable = "current_teams"
	}

	return &SQLRepository{
		db:               db,
		teamsTable:       teamsTable,
		membersTable:     membersTable,
		currentTeamTable: currentTeamTable,
		now:              time.Now,
	}
}

func (r *SQLRepository) CreateTeam(ctx context.Context, team Team) (Team, error) {
	now := r.now()
	team.CreatedAt = now
	team.UpdatedAt = now
	row := r.db.QueryRow(ctx,
		"INSERT INTO "+r.teamsTable+" (name, owner_id, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id",
		team.Name, team.OwnerID, team.CreatedAt, team.UpdatedAt,
	)

	if err := row.Scan(&team.ID); err != nil {
		return Team{}, err
	}

	if err := r.AddMember(ctx, Member{TeamID: team.ID, UserID: team.OwnerID, Role: "owner"}); err != nil {
		return Team{}, err
	}

	if err := r.SetCurrentTeam(ctx, team.OwnerID, team.ID); err != nil {
		return Team{}, err
	}

	return team, nil
}

func (r *SQLRepository) TeamsForUser(ctx context.Context, userID string) ([]Team, error) {
	rows, err := r.db.Query(ctx,
		"SELECT t.id, t.name, t.owner_id, t.created_at, t.updated_at FROM "+r.teamsTable+" t "+
			"INNER JOIN "+r.membersTable+" m ON m.team_id = t.id WHERE m.user_id = $1",
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	teams := make([]Team, 0)

	for rows.Next() {
		team, err := scanTeam(rows)

		if err != nil {
			return nil, err
		}

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

func (r *SQLRepository) FindTeam(ctx context.Context, teamID string) (Team, error) {
	row := r.db.QueryRow(ctx, "SELECT id, name, owner_id, created_at, updated_at FROM "+r.teamsTable+" WHERE id = $1 LIMIT 1", teamID)

	return scanTeam(row)
}

func (r *SQLRepository) AddMember(ctx context.Context, member Member) error {
	return r.db.Exec(ctx,
		"INSERT INTO "+r.membersTable+" (team_id, user_id, role) VALUES ($1, $2, $3) "+
			"ON CONFLICT (team_id, user_id) DO UPDATE SET role=$3",
		member.TeamID, member.UserID, member.Role,
	)
}

func (r *SQLRepository) UpdateMemberRole(ctx context.Context, teamID, userID, role string) error {
	return r.db.Exec(ctx, "UPDATE "+r.membersTable+" SET role = $1 WHERE team_id = $2 AND user_id = $3", role, teamID, userID)
}

func (r *SQLRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	if err := r.db.Exec(ctx, "DELETE FROM "+r.membersTable+" WHERE team_id = $1 AND user_id = $2", teamID, userID); err != nil {
		return err
	}

	return r.db.Exec(ctx, "DELETE FROM "+r.currentTeamTable+" WHERE user_id = $1 AND team_id = $2", userID, teamID)
}

func (r *SQLRepository) Member(ctx context.Context, teamID, userID string) (Member, error) {
	row := r.db.QueryRow(ctx, "SELECT team_id, user_id, role FROM "+r.membersTable+" WHERE team_id = $1 AND user_id = $2 LIMIT 1", teamID, userID)

	var member Member

	if err := row.Scan(&member.TeamID, &member.UserID, &member.Role); err != nil {
		return Member{}, ErrMemberNotFound
	}

	return member, nil
}

func (r *SQLRepository) SetCurrentTeam(ctx context.Context, userID, teamID string) error {
	return r.db.Exec(ctx,
		"INSERT INTO "+r.currentTeamTable+" (user_id, team_id) VALUES ($1, $2) "+
			"ON CONFLICT (user_id) DO UPDATE SET team_id=$2",
		userID, teamID,
	)
}

func (r *SQLRepository) CurrentTeam(ctx context.Context, userID string) (Team, error) {
	row := r.db.QueryRow(ctx,
		"SELECT t.id, t.name, t.owner_id, t.created_at, t.updated_at FROM "+r.teamsTable+" t "+
			"INNER JOIN "+r.currentTeamTable+" c ON c.team_id = t.id WHERE c.user_id = $1 LIMIT 1",
		userID,
	)

	return scanTeam(row)
}

func scanTeam(row interface{ Scan(dest ...any) error }) (Team, error) {
	var team Team

	if err := row.Scan(&team.ID, &team.Name, &team.OwnerID, &team.CreatedAt, &team.UpdatedAt); err != nil {
		return Team{}, ErrTeamNotFound
	}

	return team, nil
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

package teams

import (
	"context"
	"strings"
	"testing"
	"time"
)

type teamsSQLDB struct {
	id        string
	teams     []Team
	member    Member
	queries   []string
	lastQuery string
	lastExec  string
}

type teamsSQLRow struct {
	id     string
	team   Team
	member Member
	err    error
}

type teamsSQLRows struct {
	teams []Team
	pos   int
}

func TestSQLRepositoryCreatesTeamWithSafeTableFallbacks(t *testing.T) {
	db := &teamsSQLDB{id: "1"}
	repo := NewSQLRepository(db, "teams; DROP", "team_members; DROP", "current_teams; DROP")

	team, err := repo.CreateTeam(context.Background(), Team{Name: "Core", OwnerID: "owner"})

	if err != nil {
		t.Fatal(err)
	}

	if team.ID != "1" {
		t.Fatalf("team id = %q", team.ID)
	}

	if !strings.HasPrefix(db.queries[0], "INSERT INTO teams ") {
		t.Fatalf("team insert = %q", db.queries[0])
	}

	if !strings.HasPrefix(db.queries[1], "INSERT INTO team_members ") {
		t.Fatalf("member insert = %q", db.queries[1])
	}

	if !strings.HasPrefix(db.queries[2], "INSERT INTO current_teams ") {
		t.Fatalf("current team upsert = %q", db.queries[2])
	}
}

func TestSQLRepositoryListsAndFindsTeams(t *testing.T) {
	now := time.Now()
	db := &teamsSQLDB{teams: []Team{{ID: "1", Name: "Core", OwnerID: "owner", CreatedAt: now, UpdatedAt: now}}}
	repo := NewSQLRepository(db, "auth.teams", "auth.team_members", "auth.current_teams")

	found, err := repo.ForUser(context.Background(), "owner")

	if err != nil {
		t.Fatal(err)
	}

	if len(found) != 1 || found[0].ID != "1" {
		t.Fatalf("found = %#v", found)
	}

	team, err := repo.FindTeam(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if team.Name != "Core" {
		t.Fatalf("team = %#v", team)
	}

	if !strings.Contains(db.lastQuery, "FROM auth.teams") {
		t.Fatalf("query = %q", db.lastQuery)
	}
}

func TestSQLRepositoryMemberOperationsAreScoped(t *testing.T) {
	db := &teamsSQLDB{member: Member{TeamID: "1", UserID: "owner", Role: "owner"}}
	repo := NewSQLRepository(db, "teams", "team_members", "current_teams")

	member, err := repo.Member(context.Background(), "1", "owner")

	if err != nil {
		t.Fatal(err)
	}

	if member.Role != "owner" {
		t.Fatalf("member = %#v", member)
	}

	if err := repo.UpdateMemberRole(context.Background(), "1", "member", "admin"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.lastExec, "WHERE team_id = $2 AND user_id = $3") {
		t.Fatalf("update = %q", db.lastExec)
	}

	if err := repo.RemoveMember(context.Background(), "1", "member"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.lastExec, "WHERE user_id = $1 AND team_id = $2") {
		t.Fatalf("current cleanup = %q", db.lastExec)
	}
}

func (db *teamsSQLDB) QueryRow(_ context.Context, query string, args ...any) SQLRow {
	db.queries = append(db.queries, query)
	db.lastQuery = query

	if strings.HasPrefix(query, "INSERT INTO") {
		return teamsSQLRow{id: db.id}
	}

	if strings.Contains(query, "SELECT team_id") {
		return teamsSQLRow{member: db.member}
	}

	if len(db.teams) == 0 {
		return teamsSQLRow{err: ErrTeamNotFound}
	}

	return teamsSQLRow{team: db.teams[0]}
}

func (db *teamsSQLDB) Query(_ context.Context, query string, _ ...any) (SQLRows, error) {
	db.lastQuery = query

	return &teamsSQLRows{teams: db.teams}, nil
}

func (db *teamsSQLDB) Exec(_ context.Context, query string, _ ...any) error {
	db.queries = append(db.queries, query)
	db.lastExec = query

	return nil
}

func (r teamsSQLRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	if r.id != "" {
		*dest[0].(*string) = r.id

		return nil
	}

	if r.member.TeamID != "" {
		*dest[0].(*string) = r.member.TeamID
		*dest[1].(*string) = r.member.UserID
		*dest[2].(*string) = r.member.Role

		return nil
	}

	scanTeamInto(dest, r.team)

	return nil
}

func (r *teamsSQLRows) Next() bool {
	return r.pos < len(r.teams)
}

func (r *teamsSQLRows) Scan(dest ...any) error {
	scanTeamInto(dest, r.teams[r.pos])
	r.pos++

	return nil
}

func (r *teamsSQLRows) Close() error { return nil }
func (r *teamsSQLRows) Err() error   { return nil }

func scanTeamInto(dest []any, team Team) {
	*dest[0].(*string) = team.ID
	*dest[1].(*string) = team.Name
	*dest[2].(*string) = team.OwnerID
	*dest[3].(*time.Time) = team.CreatedAt
	*dest[4].(*time.Time) = team.UpdatedAt
}

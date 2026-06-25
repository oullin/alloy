package tokens

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

type tokenUser struct {
	id string
}

type tokenSQLDB struct {
	lastQuery string
	lastExec  string
	created   Token
	rows      []Token
}

type tokenSQLRow struct {
	id    string
	token Token
	err   error
}

type tokenSQLRows struct {
	rows []Token
	pos  int
}

func TestSQLRepositoryCreatesHashedTokenWithSafeTable(t *testing.T) {
	db := &tokenSQLDB{}
	repo := NewSQLRepository(db, "personal_access_tokens; DROP TABLE users")
	issuer := NewIssuer(repo)
	user := tokenUser{id: "1"}

	created, err := issuer.CreateToken(context.Background(), user, "CLI", []string{"deploy"}, nil)

	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(db.lastQuery, "INSERT INTO personal_access_tokens ") {
		t.Fatalf("query = %q", db.lastQuery)
	}

	if db.created.TokenHash == "" {
		t.Fatal("expected token hash to be stored")
	}

	if strings.Contains(created.PlainText, db.created.TokenHash) || db.created.TokenHash == created.PlainText {
		t.Fatal("SQL repository exposed or stored plaintext token")
	}

	if db.created.Name != "CLI" || db.created.UserID != "1" {
		t.Fatalf("created token = %#v", db.created)
	}

	if len(db.created.Abilities) != 1 || db.created.Abilities[0] != "deploy" {
		t.Fatalf("abilities = %#v", db.created.Abilities)
	}
}

func TestSQLRepositoryFindAndFindForUserRoundTripAbilities(t *testing.T) {
	now := time.Now()
	db := &tokenSQLDB{
		rows: []Token{{
			ID:        "1",
			UserID:    "42",
			Name:      "CLI",
			TokenHash: HashSecret("secret"),
			Abilities: []string{"read", "write"},
			CreatedAt: now,
		}},
	}
	repo := NewSQLRepository(db, "auth.personal_access_tokens")

	found, err := repo.Find(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if found.ID != "1" || len(found.Abilities) != 2 || found.Abilities[1] != "write" {
		t.Fatalf("found token = %#v", found)
	}

	tokens, err := repo.FindForUser(context.Background(), "42")

	if err != nil {
		t.Fatal(err)
	}

	if len(tokens) != 1 || tokens[0].ID != "1" {
		t.Fatalf("tokens = %#v", tokens)
	}

	if !strings.Contains(db.lastQuery, "FROM auth.personal_access_tokens") {
		t.Fatalf("query = %q", db.lastQuery)
	}
}

func TestSQLRepositoryMutationsUseUserScopedQueries(t *testing.T) {
	db := &tokenSQLDB{}
	repo := NewSQLRepository(db, "personal_access_tokens")

	if err := repo.Revoke(context.Background(), "1", "42"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.lastExec, "WHERE id = $2 AND user_id = $3") {
		t.Fatalf("revoke query = %q", db.lastExec)
	}

	if err := repo.Delete(context.Background(), "1", "42"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.lastExec, "WHERE id = $1 AND user_id = $2") {
		t.Fatalf("delete query = %q", db.lastExec)
	}

	if err := repo.Touch(context.Background(), "1"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(db.lastExec, "SET last_used_at = $1 WHERE id = $2") {
		t.Fatalf("touch query = %q", db.lastExec)
	}
}

func (u tokenUser) GetAuthIdentifierName() string { return "id" }
func (u tokenUser) GetAuthIdentifier() string     { return u.id }
func (u tokenUser) GetAuthPasswordName() string   { return "password" }
func (u tokenUser) GetAuthPassword() string       { return "" }
func (u tokenUser) SetAuthPassword(string)        {}
func (u tokenUser) GetRememberToken() string      { return "" }
func (u tokenUser) SetRememberToken(string)       {}
func (u tokenUser) GetRememberTokenName() string  { return "remember_token" }

func (db *tokenSQLDB) QueryRow(_ context.Context, query string, args ...any) SQLRow {
	db.lastQuery = query

	if strings.HasPrefix(query, "INSERT") {
		abilities := decodeAbilities(args[3].(string))
		db.created = Token{
			ID:        "1",
			UserID:    args[0].(string),
			Name:      args[1].(string),
			TokenHash: args[2].(string),
			Abilities: abilities,
			CreatedAt: args[4].(time.Time),
		}

		return tokenSQLRow{id: "1"}
	}

	if len(db.rows) == 0 {
		return tokenSQLRow{err: ErrTokenNotFound}
	}

	return tokenSQLRow{token: db.rows[0]}
}

func (db *tokenSQLDB) Query(_ context.Context, query string, _ ...any) (SQLRows, error) {
	db.lastQuery = query

	return &tokenSQLRows{rows: db.rows}, nil
}

func (db *tokenSQLDB) Exec(_ context.Context, query string, _ ...any) error {
	db.lastExec = query

	return nil
}

func (r tokenSQLRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	if r.id != "" {
		*dest[0].(*string) = r.id

		return nil
	}

	scanTokenInto(dest, r.token)

	return nil
}

func (r *tokenSQLRows) Next() bool {
	return r.pos < len(r.rows)
}

func (r *tokenSQLRows) Scan(dest ...any) error {
	scanTokenInto(dest, r.rows[r.pos])
	r.pos++

	return nil
}

func (r *tokenSQLRows) Close() error { return nil }
func (r *tokenSQLRows) Err() error   { return nil }

func scanTokenInto(dest []any, token Token) {
	abilities, _ := json.Marshal(token.Abilities)
	*dest[0].(*string) = token.ID
	*dest[1].(*string) = token.UserID
	*dest[2].(*string) = token.Name
	*dest[3].(*string) = token.TokenHash
	*dest[4].(*string) = string(abilities)
	*dest[5].(*time.Time) = token.CreatedAt
	*dest[6].(**time.Time) = cloneTime(token.LastUsedAt)
	*dest[7].(**time.Time) = cloneTime(token.ExpiresAt)
	*dest[8].(**time.Time) = cloneTime(token.RevokedAt)
}

func decodeAbilities(value string) []string {
	var abilities []string

	_ = json.Unmarshal([]byte(value), &abilities)

	return abilities
}

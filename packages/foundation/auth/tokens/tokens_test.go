package tokens_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oullin/alloy/packages/foundation/auth/httpx"
	"github.com/oullin/alloy/packages/foundation/auth/tokens"
	"github.com/oullin/alloy/packages/foundation/auth/user"
	cauth "github.com/oullin/alloy/packages/foundation/contracts/auth"
)

type stubUserProvider struct {
	users map[string]cauth.User
}

func TestIssuerStoresHashAndReturnsPlainTextOnce(t *testing.T) {
	repo := tokens.NewMemoryRepository()
	issuer := tokens.NewIssuer(repo)
	user := user.NewGenericUser(map[string]any{"id": "1"})

	created, err := issuer.CreateToken(context.Background(), user, "CLI", []string{"deploy"}, nil)

	if err != nil {
		t.Fatal(err)
	}

	if created.PlainText == "" {
		t.Fatal("expected plaintext token")
	}

	if created.AccessToken.TokenHash == "" {
		t.Fatal("expected stored token hash")
	}

	if strings.Contains(created.PlainText, created.AccessToken.TokenHash) {
		t.Fatal("plaintext response should not contain stored hash")
	}

	stored, err := repo.Find(context.Background(), created.AccessToken.ID)

	if err != nil {
		t.Fatal(err)
	}

	if stored.TokenHash == created.PlainText {
		t.Fatal("repository stored plaintext token")
	}
}

func TestFindByPlainTextTokenRejectsInvalidExpiredAndRevokedTokens(t *testing.T) {
	repo := tokens.NewMemoryRepository()
	issuer := tokens.NewIssuer(repo)
	user := user.NewGenericUser(map[string]any{"id": "1"})

	created, err := issuer.CreateToken(context.Background(), user, "CLI", []string{"deploy"}, nil)

	if err != nil {
		t.Fatal(err)
	}

	if _, err := tokens.FindByPlainTextToken(context.Background(), repo, created.PlainText); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}

	if _, err := tokens.FindByPlainTextToken(context.Background(), repo, created.AccessToken.ID+"|wrong"); err == nil {
		t.Fatal("expected invalid secret to be rejected")
	}

	if err := repo.Revoke(context.Background(), created.AccessToken.ID, user.GetAuthIdentifier()); err != nil {
		t.Fatal(err)
	}

	if _, err := tokens.FindByPlainTextToken(context.Background(), repo, created.PlainText); err != tokens.ErrTokenInactive {
		t.Fatalf("revoked token err = %v, want ErrTokenInactive", err)
	}

	expiredAt := time.Now().Add(-time.Minute)
	expired, err := issuer.CreateToken(context.Background(), user, "Expired", nil, &expiredAt)

	if err != nil {
		t.Fatal(err)
	}

	if _, err := tokens.FindByPlainTextToken(context.Background(), repo, expired.PlainText); err != tokens.ErrTokenInactive {
		t.Fatalf("expired token err = %v, want ErrTokenInactive", err)
	}
}

func TestTokenAbilities(t *testing.T) {
	token := tokens.Token{Abilities: []string{"read", "write"}}

	if !token.Can("read") {
		t.Fatal("expected token to allow read")
	}

	if token.Can("delete") {
		t.Fatal("expected token to deny delete")
	}

	wildcard := tokens.Token{Abilities: []string{"*"}}

	if !wildcard.Can("anything") {
		t.Fatal("expected wildcard token to allow any ability")
	}
}

func TestAuthenticateBearerTokenSetsUserAndCurrentToken(t *testing.T) {
	repo := tokens.NewMemoryRepository()
	issuer := tokens.NewIssuer(repo)
	user := user.NewGenericUser(map[string]any{"id": "1"})
	created, err := issuer.CreateToken(context.Background(), user, "CLI", []string{"deploy"}, nil)

	if err != nil {
		t.Fatal(err)
	}

	provider := &stubUserProvider{users: map[string]cauth.User{"1": user}}
	middleware := tokens.AuthenticateBearerToken(repo, provider)
	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		if httpx.UserFromContext(r.Context()) != user {
			t.Fatal("expected user in context")
		}

		token, ok := tokens.CurrentToken(r.Context())

		if !ok || token.ID != created.AccessToken.ID {
			t.Fatalf("current token = %#v", token)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+created.PlainText)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	if !called {
		t.Fatal("expected next handler to run")
	}

	stored, err := repo.Find(context.Background(), created.AccessToken.ID)

	if err != nil {
		t.Fatal(err)
	}

	if stored.LastUsedAt == nil {
		t.Fatal("expected token last-used timestamp to be touched")
	}
}

func TestRequireAbility(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(tokens.WithToken(req.Context(), tokens.Token{Abilities: []string{"read"}}))
	rec := httptest.NewRecorder()
	handler := tokens.RequireAbility("write")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func (p *stubUserProvider) RetrieveByID(_ context.Context, id string) (cauth.User, error) {
	return p.users[id], nil
}

func (p *stubUserProvider) RetrieveByToken(context.Context, string, string) (cauth.User, error) {
	return nil, nil
}

func (p *stubUserProvider) RetrieveByCredentials(context.Context, map[string]string) (cauth.User, error) {
	return nil, nil
}

func (p *stubUserProvider) UpdateRememberToken(context.Context, cauth.User, string) error {
	return nil
}

func (p *stubUserProvider) ValidateCredentials(context.Context, cauth.User, map[string]string) (bool, error) {
	return false, nil
}

func (p *stubUserProvider) RehashPasswordIfRequired(context.Context, cauth.User, map[string]string, bool) error {
	return nil
}

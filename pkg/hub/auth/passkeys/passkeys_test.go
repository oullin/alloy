package passkeys_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"
	"hara.sh/alloy/auth/passkeys"
	"hara.sh/alloy/auth/user"
)

func TestMemoryRepositoryCreatesStableUserHandle(t *testing.T) {
	repo := passkeys.NewMemoryRepository()

	first, err := repo.GetOrCreateUserHandle(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	second, err := repo.GetOrCreateUserHandle(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(first, second) {
		t.Fatal("expected stable user handle")
	}

	userID, err := repo.UserIDByHandle(context.Background(), first)

	if err != nil {
		t.Fatal(err)
	}

	if userID != "1" {
		t.Fatalf("userID = %q, want 1", userID)
	}
}

func TestMemoryRepositoryStoresAndUpdatesCredentials(t *testing.T) {
	repo := passkeys.NewMemoryRepository()
	credential := webauthn.Credential{ID: []byte("credential"), PublicKey: []byte("public-key")}

	if err := repo.SaveCredential(context.Background(), "1", credential); err != nil {
		t.Fatal(err)
	}

	credential.Authenticator.SignCount = 2

	if err := repo.UpdateCredential(context.Background(), "1", credential); err != nil {
		t.Fatal(err)
	}

	credentials, err := repo.CredentialsByUser(context.Background(), "1")

	if err != nil {
		t.Fatal(err)
	}

	if len(credentials) != 1 || credentials[0].Authenticator.SignCount != 2 {
		t.Fatalf("credentials = %#v", credentials)
	}
}

func TestServiceBeginsRegistrationAndStoresSessionServerSide(t *testing.T) {
	service, sessions := newTestService(t)
	user := user.NewGenericUser(map[string]any{"id": "1"})

	options, err := service.BeginRegistration(context.Background(), "session-key", user)

	if err != nil {
		t.Fatal(err)
	}

	if options == nil {
		t.Fatal("expected registration options")
	}

	session, err := sessions.Get(context.Background(), "session-key")

	if err != nil {
		t.Fatal(err)
	}

	if session.Challenge == "" {
		t.Fatal("expected challenge to be stored server-side")
	}
}

func TestServiceBeginsDiscoverableLoginAndStoresSessionServerSide(t *testing.T) {
	service, sessions := newTestService(t)

	options, err := service.BeginDiscoverableLogin(context.Background(), "login-key")

	if err != nil {
		t.Fatal(err)
	}

	if options == nil {
		t.Fatal("expected login options")
	}

	session, err := sessions.Get(context.Background(), "login-key")

	if err != nil {
		t.Fatal(err)
	}

	if session.Challenge == "" {
		t.Fatal("expected challenge to be stored server-side")
	}
}

func newTestService(t *testing.T) (*passkeys.Service, *passkeys.MemorySessionStore) {
	t.Helper()

	sessions := passkeys.NewMemorySessionStore()
	service, err := passkeys.NewService(&webauthn.Config{
		RPID:          "example.com",
		RPDisplayName: "Alloy",
		RPOrigins:     []string{"https://example.com"},
	}, passkeys.NewMemoryRepository(), sessions)

	if err != nil {
		t.Fatal(err)
	}

	return service, sessions
}

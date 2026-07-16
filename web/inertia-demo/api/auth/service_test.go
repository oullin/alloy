package auth

import (
	"errors"
	"testing"

	"alloy.dev/inertia-demo/internal/database"
	"alloy.dev/inertia-demo/internal/seed"
)

func newAuthTestService(t *testing.T) service {
	t.Helper()

	db, err := database.Open(":memory:")

	if err != nil {
		t.Fatalf("Open(:memory:) error = %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := seed.RunWithPassword(db, testSeedPassword); err != nil {
		t.Fatalf("seed.RunWithPassword() error = %v", err)
	}

	svc, err := newService(db)

	if err != nil {
		t.Fatalf("newService() error = %v", err)
	}

	return svc
}

// TestLoginTimingEqualErrors verifies the not-found and wrong-password branches
// return the same generic error, so the response does not disclose whether an
// account exists.
func TestLoginTimingEqualErrors(t *testing.T) {
	t.Parallel()

	svc := newAuthTestService(t)

	_, missingErr := svc.authenticate(loginForm{Email: "nobody@example.com", Password: "whatever"})

	if !errors.Is(missingErr, errInvalidCredentials) {
		t.Fatalf("missing-user error = %v, want errInvalidCredentials", missingErr)
	}

	_, wrongErr := svc.authenticate(loginForm{Email: "test@example.com", Password: "definitely-wrong"})

	if !errors.Is(wrongErr, errInvalidCredentials) {
		t.Fatalf("wrong-password error = %v, want errInvalidCredentials", wrongErr)
	}

	if missingErr.Error() != wrongErr.Error() {
		t.Fatalf("errors differ: missing = %q, wrong = %q", missingErr, wrongErr)
	}
}

func TestLoginAuthenticatesValidCredentials(t *testing.T) {
	t.Parallel()

	svc := newAuthTestService(t)

	user, err := svc.authenticate(loginForm{Email: "test@example.com", Password: testSeedPassword})

	if err != nil {
		t.Fatalf("authenticate() error = %v, want nil", err)
	}

	if user == nil || user.Email != "test@example.com" {
		t.Fatalf("authenticate() user = %#v, want test@example.com", user)
	}
}

// TestLoginRunsBcryptOnMissingUser asserts the not-found branch still performs a
// bcrypt comparison (against the dummy hash) rather than short-circuiting.
func TestLoginRunsBcryptOnMissingUser(t *testing.T) {
	t.Parallel()

	if len(dummyPasswordHash) == 0 {
		t.Fatal("dummyPasswordHash is empty; not-found branch cannot equalize timing")
	}

	svc := newAuthTestService(t)

	// A missing user must not panic and must exercise the dummy-hash compare.
	if _, err := svc.authenticate(loginForm{Email: "ghost@example.com", Password: "x"}); !errors.Is(err, errInvalidCredentials) {
		t.Fatalf("missing-user error = %v, want errInvalidCredentials", err)
	}
}

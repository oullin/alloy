package httpx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "alloy.dev/api/auth/httpx"
	"alloy.dev/api/auth/internal/authtest"
	"alloy.dev/api/auth/security"
	"alloy.dev/api/auth/sessionx"
	"alloy.dev/api/auth/user"
	cauth "alloy.dev/api/contracts/auth"
)

// --- EnsureAuthenticated ---

// --- RedirectIfAuthenticated ---

// --- EnsureEmailIsVerified ---

type verifiedUser struct {
	user.GenericUser
	verified bool
}

type trackingHasher struct {
	checks int
}

func (h *trackingHasher) Hash(context.Context, string) (string, error) {
	return "", nil
}

func (h *trackingHasher) Check(context.Context, string, string) (bool, error) {
	h.checks++

	return false, nil
}

func (h *trackingHasher) NeedsRehash(string) bool {
	return false
}

func TestEnsureAuthenticatedRejects(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	guard := sessionx.NewTokenGuard("api", provider)

	mw := EnsureAuthenticated(guard)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	guard.SetRequest(req)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestEnsureAuthenticatedAllows(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "api_token": "tok"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	guard := sessionx.NewTokenGuard("api", provider)

	mw := EnsureAuthenticated(guard)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")
	guard.SetRequest(req)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestEnsureAuthenticatedSetsUserInContext(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "api_token": "tok"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	guard := sessionx.NewTokenGuard("api", provider)

	mw := EnsureAuthenticated(guard)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())

		if u == nil {
			t.Error("expected user in context")
		}

		if u.GetAuthIdentifier() != "1" {
			t.Errorf("user id = %v, want \"1\"", u.GetAuthIdentifier())
		}

		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")
	guard.SetRequest(req)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRedirectIfAuthenticatedRedirectsWhenLoggedIn(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	ctx := context.Background()
	_ = guard.Login(ctx, user, false)

	mw := RedirectIfAuthenticated(guard, "/dashboard")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rr.Code)
	}

	if rr.Header().Get("Location") != "/dashboard" {
		t.Errorf("expected redirect to /dashboard, got %s", rr.Header().Get("Location"))
	}
}

func TestRedirectIfAuthenticatedPassesThroughWhenGuest(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	mw := RedirectIfAuthenticated(guard, "/dashboard")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func (u *verifiedUser) HasVerifiedEmail() bool          { return u.verified }
func (u *verifiedUser) MarkEmailAsVerified(_ time.Time) {}
func (u *verifiedUser) MarkEmailAsUnverified()          {}
func (u *verifiedUser) GetEmailForVerification() string { return "test@example.com" }

func TestEnsureEmailIsVerifiedRejects(t *testing.T) {
	user := &verifiedUser{
		GenericUser: *user.NewGenericUser(map[string]any{"id": "1", "password": "pw"}),
		verified:    false,
	}
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	_ = guard.Login(context.Background(), user, false)

	mw := EnsureEmailIsVerified(guard)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestEnsureEmailIsVerifiedAllowsVerifiedUser(t *testing.T) {
	user := &verifiedUser{
		GenericUser: *user.NewGenericUser(map[string]any{"id": "1", "password": "pw"}),
		verified:    true,
	}
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	_ = guard.Login(context.Background(), user, false)

	mw := EnsureEmailIsVerified(guard)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestEnsureEmailIsVerifiedAllowsUserWithoutMustVerifyEmail(t *testing.T) {
	// User doesn't implement MustVerifyEmail — should pass through.
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	_ = guard.Login(context.Background(), user, false)

	mw := EnsureEmailIsVerified(guard)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestEnsureEmailIsVerifiedRejectsUnauthenticatedUser(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	mw := EnsureEmailIsVerified(guard)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// --- RequirePassword ---

func TestRequirePasswordRedirectsWhenNotConfirmed(t *testing.T) {
	sess := authtest.NewSession()

	mw := RequirePassword(sess, 30*time.Minute, "/confirm-password")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rr.Code)
	}

	if rr.Header().Get("Location") != "/confirm-password" {
		t.Errorf("expected redirect to /confirm-password, got %s", rr.Header().Get("Location"))
	}
}

func TestRequirePasswordAllowsRecentlyConfirmed(t *testing.T) {
	sess := authtest.NewSession()
	sess.Put("auth.password_confirmed_at", time.Now().Unix())

	mw := RequirePassword(sess, 30*time.Minute, "/confirm-password")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRequirePasswordAcceptsJSONNumericTimestamp(t *testing.T) {
	sess := authtest.NewSession()
	sess.Put("auth.password_confirmed_at", float64(time.Now().Unix()))

	mw := RequirePassword(sess, 30*time.Minute, "/confirm-password")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRequirePasswordRedirectsWhenExpired(t *testing.T) {
	sess := authtest.NewSession()
	sess.Put("auth.password_confirmed_at", time.Now().Add(-1*time.Hour).Unix())

	mw := RequirePassword(sess, 30*time.Minute, "/confirm-password")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rr.Code)
	}
}

func TestRequirePasswordAcceptsIntTimestamp(t *testing.T) {
	sess := authtest.NewSession()
	sess.Put("auth.password_confirmed_at", int(time.Now().Unix()))

	mw := RequirePassword(sess, 30*time.Minute, "/confirm-password")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// --- WithBasicAuth ---

func TestAuthenticateWithBasicAuthRejectsNoCredentials(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	hasher := security.NewBcryptHasher(0)

	mw := WithBasicAuth(provider, hasher)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}

	if rr.Header().Get("WWW-Authenticate") == "" {
		t.Error("expected WWW-Authenticate header")
	}
}

func TestAuthenticateWithBasicAuthAcceptsValid(t *testing.T) {
	hasher := security.NewBcryptHasher(4)
	ctx := context.Background()
	hash, _ := hasher.Hash(ctx, "secret")
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "user@test.com", "password": hash})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}

	mw := WithBasicAuth(provider, hasher)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())

		if u == nil {
			t.Error("expected user in context after basic auth")
		}

		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("user@test.com", "secret")
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAuthenticateWithBasicAuthRejectsWrongPassword(t *testing.T) {
	hasher := security.NewBcryptHasher(4)
	ctx := context.Background()
	hash, _ := hasher.Hash(ctx, "secret")
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "user@test.com", "password": hash})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}

	mw := WithBasicAuth(provider, hasher)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("user@test.com", "wrong")
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthenticateWithBasicAuthChecksDummyPasswordWhenUserMissing(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	hasher := &trackingHasher{}

	mw := WithBasicAuth(provider, hasher)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("missing@test.com", "secret")
	mw(inner).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}

	if hasher.checks != 1 {
		t.Fatalf("hasher checks = %d, want 1", hasher.checks)
	}
}

// --- WithUser / UserFromContext ---

func TestWithUserAndUserFromContext(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = WithUser(req, user)

	got := UserFromContext(req.Context())

	if got == nil {
		t.Fatal("expected user from context")
	}

	if got.GetAuthIdentifier() != "1" {
		t.Errorf("user id = %v, want \"1\"", got.GetAuthIdentifier())
	}
}

func TestUserFromContextReturnsNilWhenNotSet(t *testing.T) {
	got := UserFromContext(context.Background())

	if got != nil {
		t.Error("expected nil when no user in context")
	}
}

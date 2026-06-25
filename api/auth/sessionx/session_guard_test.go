package sessionx_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oullin/alloy/auth/events"
	"github.com/oullin/alloy/auth/internal/authtest"
	"github.com/oullin/alloy/auth/security"
	"github.com/oullin/alloy/auth/sessionx"
	"github.com/oullin/alloy/auth/user"
	cauth "github.com/oullin/alloy/contracts/auth"
)

// --- SessionGuard: User resolution ---

func TestSessionGuardUserReturnsNilWhenNoUserFound(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	u, err := guard.User(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if u != nil {
		t.Error("expected nil user when no session and no cookie")
	}
}

func TestSessionGuardUserReturnsCachedUser(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	ctx := context.Background()

	_ = guard.Login(ctx, user, false)

	// Calling User() multiple times should return the same cached user.
	u1, _ := guard.User(ctx)
	u2, _ := guard.User(ctx)

	if u1 != u2 {
		t.Error("User() should return cached user on subsequent calls")
	}
}

func TestSessionGuardUserIsSetToRetrievedUser(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	sess.Put("_auth_user", "1")
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	got, err := guard.User(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if got != user {
		t.Error("expected user from session to match provider user")
	}
}

func TestSessionGuardUserReturnsRetrieveByIDError(t *testing.T) {
	provider := &authtest.Provider{
		Users:           map[string]cauth.User{},
		RetrieveByIDErr: authtest.ErrProvider,
	}
	sess := authtest.NewSession()
	sess.Put("_auth_user", "1")
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	_, err := guard.User(context.Background())

	if !errors.Is(err, authtest.ErrProvider) {
		t.Fatalf("User error = %v, want %v", err, authtest.ErrProvider)
	}
}

func TestSessionGuardUserUsesRememberCookieIfItExists(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	user.SetRememberToken("recaller")
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "web_remember", Value: "1|recaller|" + sessionx.HashPasswordForCookie("pw")})
	guard.SetRequest(req)

	got, err := guard.User(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if got == nil {
		t.Fatal("expected user from remember cookie")
	}

	if !guard.ViaRemember(context.Background()) {
		t.Error("expected ViaRemember() to be true")
	}
}

func TestSessionGuardUserReturnsRetrieveByTokenError(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	user.SetRememberToken("recaller")
	provider := &authtest.Provider{
		Users:              map[string]cauth.User{"1": user},
		RetrieveByTokenErr: authtest.ErrProvider,
	}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "web_remember", Value: "1|recaller|" + sessionx.HashPasswordForCookie("pw")})
	guard.SetRequest(req)

	_, err := guard.User(context.Background())

	if !errors.Is(err, authtest.ErrProvider) {
		t.Fatalf("User error = %v, want %v", err, authtest.ErrProvider)
	}
}

func TestSessionGuardUserDispatchesAuthenticatedEventFromSession(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	sess.Put("_auth_user", "1")
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	_, _ = guard.User(context.Background())

	dispatcher.Has(t, "events.Authenticated")
}

func TestSessionGuardUserDispatchesLoginEventFromRememberCookie(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	user.SetRememberToken("tok")
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "web_remember", Value: "1|tok|" + sessionx.HashPasswordForCookie("pw")})
	guard.SetRequest(req)

	_, _ = guard.User(context.Background())

	dispatcher.Has(t, "events.Login")
}

// --- SessionGuard: Check / Guest / ID ---

func TestSessionGuardCheckReturnsTrueWhenUserIsNotNull(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	ctx := context.Background()

	_ = guard.Login(ctx, user, false)

	if !guard.Check(ctx) {
		t.Error("Check() should return true after login")
	}

	if guard.Guest(ctx) {
		t.Error("Guest() should return false after login")
	}
}

func TestSessionGuardCheckReturnsFalseWhenUserIsNull(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	ctx := context.Background()

	if guard.Check(ctx) {
		t.Error("Check() should return false with no user")
	}

	if !guard.Guest(ctx) {
		t.Error("Guest() should return true with no user")
	}
}

func TestSessionGuardIDReturnsIdentifier(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "42", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"42": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	ctx := context.Background()

	_ = guard.Login(ctx, user, false)

	if guard.ID(ctx) != "42" {
		t.Errorf("ID() = %v, want \"42\"", guard.ID(ctx))
	}
}

func TestSessionGuardIDReturnsNilWhenNoUser(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	if guard.ID(context.Background()) != nil {
		t.Error("ID() should return nil when no user")
	}
}

// --- SessionGuard: Attempt ---

func TestSessionGuardAttemptCallsRetrieveByCredentials(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	ok := guard.Attempt(context.Background(), map[string]string{"email": "foo@bar.com", "password": "secret"}, false)

	if ok {
		t.Error("Attempt should fail when user not found")
	}

	dispatcher.Has(t, "events.Attempting")
	dispatcher.Has(t, "events.Failed")
	dispatcher.HasNot(t, "events.Validated")
}

func TestSessionGuardAttemptReturnsTrue(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	ok := guard.Attempt(context.Background(), map[string]string{"email": "a@b.com", "password": "pw"}, false)

	if !ok {
		t.Error("Attempt should return true with valid credentials")
	}

	dispatcher.Has(t, "events.Attempting")
	dispatcher.Has(t, "events.Validated")
	dispatcher.Has(t, "events.Login")
	dispatcher.HasNot(t, "events.Failed")
}

func TestSessionGuardAttemptReturnsFalseWithInvalidPassword(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	ok := guard.Attempt(context.Background(), map[string]string{"email": "a@b.com", "password": "wrong"}, false)

	if ok {
		t.Error("Attempt should return false with invalid password")
	}

	dispatcher.Has(t, "events.Attempting")
	dispatcher.Has(t, "events.Failed")
	dispatcher.HasNot(t, "events.Validated")
	dispatcher.HasNot(t, "events.Login")
}

func TestSessionGuardAttemptReturnsFalseIfUserNotFound(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	ok := guard.Attempt(context.Background(), map[string]string{"email": "unknown@b.com", "password": "pw"}, false)

	if ok {
		t.Error("Attempt should return false when user not found")
	}

	dispatcher.Has(t, "events.Attempting")
	dispatcher.Has(t, "events.Failed")
}

func TestSessionGuardAttemptWithRememberSetsRememberCookie(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)

	ok := guard.Attempt(context.Background(), map[string]string{"email": "a@b.com", "password": "pw"}, true)

	if !ok {
		t.Fatal("Attempt with remember should succeed")
	}

	if len(cookies.Queued) == 0 {
		t.Error("expected remember cookie to be queued")
	}
}

func TestSessionGuardFailedEventContainsUserWhenPasswordInvalid(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	guard.Attempt(context.Background(), map[string]string{"email": "a@b.com", "password": "wrong"}, false)

	dispatcher.Mu.Lock()

	defer dispatcher.Mu.Unlock()

	for _, e := range dispatcher.Events {
		if fe, ok := e.(events.Failed); ok {
			if fe.User == nil {
				t.Error("Failed event should contain the user when password is invalid")
			}

			if fe.Guard != "web" {
				t.Errorf("Failed.Guard = %q, want %q", fe.Guard, "web")
			}

			return
		}
	}

	t.Error("Failed event not dispatched")
}

func TestSessionGuardFailedEventHasNilUserWhenNotFound(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	guard.Attempt(context.Background(), map[string]string{"email": "missing@b.com", "password": "pw"}, false)

	dispatcher.Mu.Lock()

	defer dispatcher.Mu.Unlock()

	for _, e := range dispatcher.Events {
		if fe, ok := e.(events.Failed); ok {
			if fe.User != nil {
				t.Error("Failed event should have nil user when user not found")
			}

			return
		}
	}

	t.Error("Failed event not dispatched")
}

// --- SessionGuard: Login ---

func TestSessionGuardLoginStoresIdentifierInSession(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "42", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"42": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	_ = guard.Login(context.Background(), user, false)

	if sess.Get("_auth_user", nil) != "42" {
		t.Errorf("session _auth_user = %v, want \"42\"", sess.Get("_auth_user", nil))
	}
}

func TestSessionGuardLoginMigratesSessionBeforeStoringUser(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "42", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"42": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	_ = guard.Login(context.Background(), user, false)

	if sess.MigrateCalls != 1 {
		t.Fatalf("session migrate calls = %d, want 1", sess.MigrateCalls)
	}

	if len(sess.MigrateDestroy) != 1 || !sess.MigrateDestroy[0] {
		t.Fatalf("session migrate destroy flags = %v, want [true]", sess.MigrateDestroy)
	}
}

func TestSessionGuardLoginFiresLoginEvent(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	_ = guard.Login(context.Background(), user, false)

	dispatcher.Has(t, "events.Login")
}

func TestSessionGuardLoginFiresLoginEventWithRemember(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	_ = guard.Login(context.Background(), user, true)

	dispatcher.Mu.Lock()

	defer dispatcher.Mu.Unlock()

	for _, e := range dispatcher.Events {
		if le, ok := e.(events.Login); ok {
			if !le.Remember {
				t.Error("Login.Remember = false, want true")
			}

			return
		}
	}

	t.Error("Login event not found")
}

func TestSessionGuardLoginQueuesCookieWhenRemembering(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)

	_ = guard.Login(context.Background(), user, true)

	if len(cookies.Queued) == 0 {
		t.Error("expected remember cookie to be queued")
	}

	if cookies.Queued[0].Name != "web_remember" {
		t.Errorf("cookie name = %q, want %q", cookies.Queued[0].Name, "web_remember")
	}

	if !cookies.Queued[0].HttpOnly {
		t.Error("remember cookie should be HttpOnly")
	}

	if cookies.Queued[0].SameSite != http.SameSiteLaxMode {
		t.Errorf("remember cookie SameSite = %v, want Lax", cookies.Queued[0].SameSite)
	}
}

func TestSessionGuardLoginCreatesRememberTokenIfOneDoesNotExist(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)

	if user.GetRememberToken() != "" {
		t.Fatal("expected no remember token initially")
	}

	_ = guard.Login(context.Background(), user, true)

	if user.GetRememberToken() == "" {
		t.Error("expected remember token to be set after login with remember")
	}
}

// --- SessionGuard: LoginUsingID ---

func TestSessionGuardLoginUsingIDLogsInWithUser(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "10", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"10": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	got, err := guard.LoginUsingID(context.Background(), "10", false)

	if err != nil {
		t.Fatal(err)
	}

	if got != user {
		t.Error("expected returned user to match")
	}

	if !guard.Check(context.Background()) {
		t.Error("user should be authenticated after LoginUsingID")
	}
}

func TestSessionGuardLoginUsingIDFailure(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	_, err := guard.LoginUsingID(context.Background(), "11", false)

	if err == nil {
		t.Error("expected error for missing user")
	}
}

// --- SessionGuard: OnceUsingID ---

func TestSessionGuardOnceUsingIDSetsUser(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "10", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"10": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	got, err := guard.OnceUsingID(context.Background(), "10")

	if err != nil {
		t.Fatal(err)
	}

	if got != user {
		t.Error("expected returned user to match")
	}

	if guard.Check(context.Background()) != true {
		t.Error("user should be authenticated for this request")
	}

	// Should NOT have stored in session.
	if sess.Get("_auth_user", nil) != nil {
		t.Error("OnceUsingID should not persist to session")
	}
}

func TestSessionGuardOnceUsingIDFailure(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	_, err := guard.OnceUsingID(context.Background(), "11")

	if err == nil {
		t.Error("expected error for missing user")
	}
}

// --- SessionGuard: Once ---

func TestSessionGuardOnceSetsUserWithoutSession(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	ok := guard.Once(context.Background(), map[string]string{"email": "a@b.com", "password": "pw"})

	if !ok {
		t.Error("Once should return true for valid credentials")
	}

	if !guard.Check(context.Background()) {
		t.Error("should be authenticated after Once")
	}

	// Should NOT persist in session.
	if sess.Get("_auth_user", nil) != nil {
		t.Error("Once should not persist to session")
	}
}

func TestSessionGuardOnceFailure(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	ok := guard.Once(context.Background(), map[string]string{"email": "a@b.com", "password": "wrong"})

	if ok {
		t.Error("Once should return false for invalid credentials")
	}
}

// --- SessionGuard: Validate ---

func TestSessionGuardValidateReturnsTrue(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	ok := guard.Validate(context.Background(), map[string]string{"email": "a@b.com", "password": "pw"})

	if !ok {
		t.Error("Validate should return true for valid credentials")
	}

	// Should NOT log in.
	if guard.Check(context.Background()) {
		t.Error("Validate should not log in the user")
	}
}

func TestSessionGuardValidateReturnsFalse(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	ok := guard.Validate(context.Background(), map[string]string{"email": "a@b.com", "password": "wrong"})

	if ok {
		t.Error("Validate should return false for invalid credentials")
	}
}

// --- SessionGuard: Logout ---

func TestSessionGuardLogoutRemovesSessionAndCookie(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)
	ctx := context.Background()

	_ = guard.Login(ctx, user, true)

	if sess.Get("_auth_user", nil) == nil {
		t.Fatal("session should have user after login")
	}

	_ = guard.Logout(ctx)

	if sess.Get("_auth_user", nil) != nil {
		t.Error("session should be cleared after logout")
	}

	if guard.Check(ctx) {
		t.Error("should not be authenticated after logout")
	}

	// Should have a forget cookie queued.
	if len(cookies.Forgotten) == 0 {
		t.Error("expected remember cookie to be forgotten")
	}
}

func TestSessionGuardLogoutFiresLogoutEvent(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)
	ctx := context.Background()

	_ = guard.Login(ctx, user, false)
	dispatcher.Events = nil

	_ = guard.Logout(ctx)
	dispatcher.Has(t, "events.Logout")
}

func TestSessionGuardLogoutDoesNotFireEventIfNoUser(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	_ = guard.Logout(context.Background())

	dispatcher.HasNot(t, "events.Logout")
}

func TestSessionGuardLogoutRefreshesRememberToken(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	user.SetRememberToken("old-token")
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	ctx := context.Background()

	_ = guard.Login(ctx, user, false)
	_ = guard.Logout(ctx)

	if user.GetRememberToken() == "old-token" {
		t.Error("remember token should be refreshed on logout")
	}
}

// --- SessionGuard: LogoutOtherDevices ---

func TestSessionGuardLogoutOtherDevicesDispatchesEvent(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)
	ctx := context.Background()

	_ = guard.Login(ctx, user, false)
	dispatcher.Events = nil

	err := guard.LogoutOtherDevices(ctx)

	if err != nil {
		t.Fatal(err)
	}

	dispatcher.Has(t, "events.OtherDeviceLogout")
}

// --- SessionGuard: ViaRemember ---

func TestSessionGuardViaRememberReturnsFalseByDefault(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	ctx := context.Background()

	_ = guard.Login(ctx, user, false)

	if guard.ViaRemember(ctx) {
		t.Error("ViaRemember should be false when not authenticated via cookie")
	}
}

// --- SessionGuard: SetUser / HasUser / ForgetUser ---

func TestSessionGuardSetUserFiresAuthenticatedEvent(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	guard.SetUser(context.Background(), user)

	dispatcher.Has(t, "events.Authenticated")
}

func TestSessionGuardHasUserReturnsTrueWhenUserIsSet(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	if guard.HasUser() {
		t.Error("HasUser should be false initially")
	}

	guard.SetUser(context.Background(), user)

	if !guard.HasUser() {
		t.Error("HasUser should be true after SetUser")
	}
}

func TestSessionGuardForgetUserClearsUser(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	guard.SetUser(context.Background(), user)
	guard.ForgetUser()

	if guard.HasUser() {
		t.Error("HasUser should be false after ForgetUser")
	}
}

// --- SessionGuard: LogoutCurrentDevice ---

func TestSessionGuardLogoutCurrentDeviceFiresEvent(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)
	ctx := context.Background()

	_ = guard.Login(ctx, user, false)
	dispatcher.Events = nil

	err := guard.LogoutCurrentDevice(ctx)

	if err != nil {
		t.Fatal(err)
	}

	dispatcher.Has(t, "events.CurrentDeviceLogout")
	dispatcher.HasNot(t, "events.Logout")
}

func TestSessionGuardLogoutCurrentDeviceRemovesSessionAndCookie(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)
	ctx := context.Background()

	_ = guard.Login(ctx, user, true)
	_ = guard.LogoutCurrentDevice(ctx)

	if sess.Get("_auth_user", nil) != nil {
		t.Error("session should be cleared")
	}

	if guard.HasUser() {
		t.Error("user should be cleared")
	}

	if len(cookies.Forgotten) == 0 {
		t.Error("remember cookie should be forgotten")
	}
}

func TestSessionGuardLogoutCurrentDeviceDoesNotRefreshRememberToken(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	user.SetRememberToken("original-token")
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	ctx := context.Background()

	_ = guard.Login(ctx, user, false)

	// Reset token back after login may have changed it.
	user.SetRememberToken("original-token")

	_ = guard.LogoutCurrentDevice(ctx)

	if user.GetRememberToken() != "original-token" {
		t.Error("LogoutCurrentDevice should NOT refresh remember token (unlike Logout)")
	}
}

// --- SessionGuard: Login fires both Login and Authenticated ---

func TestSessionGuardLoginFiresBothLoginAndAuthenticatedEvents(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	_ = guard.Login(context.Background(), user, false)

	dispatcher.Has(t, "events.Login")
	dispatcher.Has(t, "events.Authenticated")
}

// --- SessionGuard: No dispatcher (backward compat) ---

func TestSessionGuardWorksWithoutDispatcher(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	ctx := context.Background()

	// All operations should work without panic.
	ok := guard.Attempt(ctx, map[string]string{"email": "a@b.com", "password": "pw"}, false)

	if !ok {
		t.Fatal("Attempt should succeed without dispatcher")
	}

	_ = guard.Logout(ctx)
}

// --- SessionGuard: GetLastAttempted ---

func TestSessionGuardGetLastAttemptedReturnsNilInitially(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	if guard.GetLastAttempted() != nil {
		t.Error("GetLastAttempted should be nil initially")
	}
}

func TestSessionGuardGetLastAttemptedReturnsUserAfterFailedAttempt(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	guard.Attempt(context.Background(), map[string]string{"email": "a@b.com", "password": "wrong"}, false)

	if guard.GetLastAttempted() != user {
		t.Error("GetLastAttempted should return the user even after failed attempt")
	}
}

func TestSessionGuardGetLastAttemptedReturnsUserAfterSuccessfulAttempt(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	guard.Attempt(context.Background(), map[string]string{"email": "a@b.com", "password": "pw"}, false)

	if guard.GetLastAttempted() != user {
		t.Error("GetLastAttempted should return the user after successful attempt")
	}
}

// --- SessionGuard: AttemptWhen ---

func TestSessionGuardAttemptWhenSucceedsWhenAllCallbacksPass(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	callbacks := []func(cauth.User) bool{
		func(u cauth.User) bool { return u != nil },
		func(u cauth.User) bool { return u.GetAuthIdentifier() == "1" },
	}

	ok := guard.AttemptWhen(context.Background(), map[string]string{"email": "a@b.com", "password": "pw"}, callbacks, false)

	if !ok {
		t.Error("AttemptWhen should succeed when all callbacks pass")
	}

	if !guard.Check(context.Background()) {
		t.Error("user should be logged in after AttemptWhen")
	}
}

func TestSessionGuardAttemptWhenFailsWhenCallbackReturnsFalse(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	callbacks := []func(cauth.User) bool{
		func(_ cauth.User) bool { return false },
	}

	ok := guard.AttemptWhen(context.Background(), map[string]string{"email": "a@b.com", "password": "pw"}, callbacks, false)

	if ok {
		t.Error("AttemptWhen should fail when callback returns false")
	}

	if guard.Check(context.Background()) {
		t.Error("user should not be logged in")
	}

	dispatcher.Has(t, "events.Failed")
}

func TestSessionGuardAttemptWhenFailsWithInvalidCredentials(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	callbacks := []func(cauth.User) bool{
		func(_ cauth.User) bool { return true },
	}

	ok := guard.AttemptWhen(context.Background(), map[string]string{"email": "a@b.com", "password": "wrong"}, callbacks, false)

	if ok {
		t.Error("AttemptWhen should fail with invalid credentials")
	}
}

// --- SessionGuard: Basic / OnceBasic ---

func TestSessionGuardImplementsSupportsBasicAuth(t *testing.T) {
	var _ cauth.SupportsBasicAuth = (*sessionx.SessionGuard)(nil)
}

func TestSessionGuardBasicAuthenticatesFromRequest(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("a@b.com", "pw")
	guard.SetRequest(req)

	ok := guard.Basic(context.Background(), "email", nil)

	if !ok {
		t.Error("Basic should succeed with valid credentials")
	}

	if !guard.Check(context.Background()) {
		t.Error("user should be authenticated after Basic")
	}
}

func TestSessionGuardBasicReturnsFalseWithInvalidCredentials(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("a@b.com", "wrong")
	guard.SetRequest(req)

	ok := guard.Basic(context.Background(), "email", nil)

	if ok {
		t.Error("Basic should fail with invalid credentials")
	}
}

func TestSessionGuardBasicWithExtraConditions(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw", "active": "1"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("a@b.com", "pw")
	guard.SetRequest(req)

	ok := guard.Basic(context.Background(), "email", map[string]string{"active": "1"})

	if !ok {
		t.Error("Basic should include extra credential conditions")
	}
}

func TestSessionGuardBasicReturnsTrueIfAlreadyAuthenticated(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	_ = guard.Login(context.Background(), user, false)

	ok := guard.Basic(context.Background(), "email", nil)

	if !ok {
		t.Error("Basic should return true when already authenticated")
	}
}

func TestSessionGuardOnceBasicAuthenticatesWithoutSession(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "email": "a@b.com", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("a@b.com", "pw")
	guard.SetRequest(req)

	ok := guard.OnceBasic(context.Background(), "email", nil)

	if !ok {
		t.Error("OnceBasic should succeed with valid credentials")
	}

	if sess.Get("_auth_user", nil) != nil {
		t.Error("OnceBasic should not persist to session")
	}
}

// --- SessionGuard: HashPasswordForCookie ---

func TestHashPasswordForCookie(t *testing.T) {
	hash := sessionx.HashPasswordForCookie("secret")

	if len(hash) != 10 {
		t.Errorf("HashPasswordForCookie should return 10 chars, got %d", len(hash))
	}

	// Same input should produce same hash.
	if sessionx.HashPasswordForCookie("secret") != hash {
		t.Error("HashPasswordForCookie should be deterministic")
	}

	// Different input should produce different hash.
	if sessionx.HashPasswordForCookie("other") == hash {
		t.Error("different passwords should produce different hashes")
	}
}

func TestSessionGuardLoginCookieContainsPasswordHash(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)

	_ = guard.Login(context.Background(), user, true)

	if len(cookies.Queued) == 0 {
		t.Fatal("expected remember cookie")
	}

	recaller := security.NewRecaller(cookies.Queued[0].Value)

	if recaller == nil {
		t.Fatal("expected valid recaller from cookie")
	}

	if !recaller.Valid() {
		t.Error("recaller should be valid (all three parts non-empty)")
	}

	if recaller.Hash() != sessionx.HashPasswordForCookie("pw") {
		t.Error("cookie hash should match HashPasswordForCookie of user password")
	}
}

// --- SessionGuard: Accessors ---

func TestSessionGuardAccessors(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)

	if guard.GetName() != "web" {
		t.Errorf("GetName() = %q, want %q", guard.GetName(), "web")
	}

	if guard.GetRecallerName() != "web_remember" {
		t.Errorf("GetRecallerName() = %q, want %q", guard.GetRecallerName(), "web_remember")
	}

	if guard.GetCookieJar() != cookies {
		t.Error("GetCookieJar should return the cookies")
	}

	if guard.GetSession() != sess {
		t.Error("GetSession should return the session")
	}

	if guard.GetUser() != nil {
		t.Error("GetUser should be nil before login")
	}

	if guard.GetProvider() != provider {
		t.Error("GetProvider should return the provider")
	}

	if guard.GetRequest() != nil {
		t.Error("GetRequest should be nil before SetRequest")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	guard.SetRequest(req)

	if guard.GetRequest() != req {
		t.Error("GetRequest should return the set request")
	}

	_ = guard.Login(context.Background(), user, false)

	if guard.GetUser() != user {
		t.Error("GetUser should return user after login")
	}
}

func TestSessionGuardSetProvider(t *testing.T) {
	provider1 := &authtest.Provider{Users: map[string]cauth.User{}}
	provider2 := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider1, sess, nil, nil)

	guard.SetProvider(provider2)

	if guard.GetProvider() != provider2 {
		t.Error("SetProvider should update the provider")
	}
}

func TestSessionGuardSetCookieJar(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)
	cookies := &authtest.CookieManager{}

	guard.SetCookieJar(cookies)

	if guard.GetCookieJar() != cookies {
		t.Error("SetCookieJar should update the cookie jar")
	}
}

func TestSessionGuardGetDispatcher(t *testing.T) {
	provider := &authtest.Provider{Users: map[string]cauth.User{}}
	sess := authtest.NewSession()
	guard := sessionx.NewSessionGuard("web", provider, sess, nil, nil)

	if guard.GetDispatcher() != nil {
		t.Error("GetDispatcher should be nil before SetEventDispatcher")
	}

	dispatcher := &authtest.Dispatcher{}
	guard.SetEventDispatcher(dispatcher)

	if guard.GetDispatcher() != dispatcher {
		t.Error("GetDispatcher should return the set dispatcher")
	}
}

func TestSessionGuardSetRememberDuration(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)

	guard.SetRememberDuration(60) // 60 minutes

	_ = guard.Login(context.Background(), user, true)

	if len(cookies.Queued) == 0 {
		t.Fatal("expected remember cookie")
	}

	if cookies.Queued[0].MaxAge != 3600 {
		t.Errorf("MaxAge = %d, want 3600", cookies.Queued[0].MaxAge)
	}
}

func TestSessionGuardSetRememberCookie(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "pw"})
	provider := &authtest.Provider{Users: map[string]cauth.User{"1": user}}
	sess := authtest.NewSession()
	cookies := &authtest.CookieManager{}
	guard := sessionx.NewSessionGuard("web", provider, sess, cookies, nil)

	guard.SetRememberCookie(http.Cookie{
		Path:     "/admin",
		Domain:   "example.com",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	_ = guard.Login(context.Background(), user, true)

	if len(cookies.Queued) == 0 {
		t.Fatal("expected remember cookie")
	}

	cookie := cookies.Queued[0]

	if cookie.Path != "/admin" {
		t.Errorf("Path = %q, want /admin", cookie.Path)
	}

	if cookie.Domain != "example.com" {
		t.Errorf("Domain = %q, want example.com", cookie.Domain)
	}

	if !cookie.Secure {
		t.Error("Secure = false, want true")
	}

	if !cookie.HttpOnly {
		t.Error("HttpOnly should remain true")
	}

	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("SameSite = %v, want Strict", cookie.SameSite)
	}
}

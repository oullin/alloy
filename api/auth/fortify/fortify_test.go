package fortify_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	cauth "github.com/oullin/alloy/api/contracts/auth"
	"github.com/oullin/alloy/auth/browserx"
	"github.com/oullin/alloy/auth/fortify"
	"github.com/oullin/alloy/auth/passkeys"
	"github.com/oullin/alloy/auth/passwords"
	"github.com/oullin/alloy/auth/teams"
	patokens "github.com/oullin/alloy/auth/tokens"
	"github.com/oullin/alloy/auth/twofactor"
	"github.com/oullin/alloy/auth/user"
)

type stubStatefulGuard struct {
	user               cauth.User
	attemptOK          bool
	attemptCredentials map[string]string
	attemptRemember    bool
	loginUser          cauth.User
	loggedOut          bool
}

type stubResetLinkSender struct {
	email string
	err   error
}

type stubPasswordResetter struct {
	credentials map[string]any
	err         error
}

type resetUser struct {
	email string
}

type verifiableUser struct {
	*user.GenericUser
	verified bool
	notified bool
}

type twoFactorUser struct {
	*user.GenericUser
	enabled       bool
	secret        string
	recoveryCodes []string
	confirmedAt   *time.Time
}

type stubHasher struct {
	checkOK bool
	hash    string
}

type stubConfirmationSession struct {
	values map[string]any
}

func TestRoutesReturnsEnabledHeadlessRouteContracts(t *testing.T) {
	handler := func(http.ResponseWriter, *http.Request) {}

	routes := fortify.Routes(fortify.Actions{
		Login:          handler,
		Logout:         handler,
		UpdatePassword: handler,
	})

	if len(routes) != 3 {
		t.Fatalf("len(routes) = %d, want 3", len(routes))
	}

	assertRoute(t, routes[0], http.MethodPost, "/login", "login", []string{"guest"})
	assertRoute(t, routes[1], http.MethodPost, "/logout", "logout", []string{"auth"})
	assertRoute(t, routes[2], http.MethodPut, "/user/password", "user-password.update", []string{"auth"})
}

func TestRoutesIncludesEnabledAPITokenContracts(t *testing.T) {
	handler := func(http.ResponseWriter, *http.Request) {}

	routes := fortify.Routes(fortify.Actions{
		ListAPITokens:  handler,
		CreateAPIToken: handler,
		RevokeAPIToken: handler,
	})

	if len(routes) != 3 {
		t.Fatalf("len(routes) = %d, want 3", len(routes))
	}

	assertRoute(t, routes[0], http.MethodGet, "/user/api-tokens", "api-tokens.index", []string{"auth"})
	assertRoute(t, routes[1], http.MethodPost, "/user/api-tokens", "api-tokens.store", []string{"auth"})
	assertRoute(t, routes[2], http.MethodDelete, "/user/api-tokens/{token}", "api-tokens.destroy", []string{"auth"})
}

func TestRoutesIncludesEnabledTwoFactorContracts(t *testing.T) {
	handler := func(http.ResponseWriter, *http.Request) {}

	routes := fortify.Routes(fortify.Actions{
		EnableTwoFactor:         handler,
		ConfirmTwoFactor:        handler,
		DisableTwoFactor:        handler,
		RegenerateRecoveryCodes: handler,
	})

	if len(routes) != 4 {
		t.Fatalf("len(routes) = %d, want 4", len(routes))
	}

	assertRoute(t, routes[0], http.MethodPost, "/user/two-factor-authentication", "two-factor.enable", []string{"auth", "password.confirm"})
	assertRoute(t, routes[1], http.MethodPost, "/user/confirmed-two-factor-authentication", "two-factor.confirm", []string{"auth", "password.confirm"})
	assertRoute(t, routes[2], http.MethodDelete, "/user/two-factor-authentication", "two-factor.disable", []string{"auth", "password.confirm"})
	assertRoute(t, routes[3], http.MethodPost, "/user/two-factor-recovery-codes", "two-factor.recovery-codes", []string{"auth", "password.confirm"})
}

func TestRoutesIncludesEnabledBrowserSessionContracts(t *testing.T) {
	handler := func(http.ResponseWriter, *http.Request) {}

	routes := fortify.Routes(fortify.Actions{
		ListBrowserSessions:        handler,
		RevokeBrowserSession:       handler,
		RevokeOtherBrowserSessions: handler,
	})

	if len(routes) != 3 {
		t.Fatalf("len(routes) = %d, want 3", len(routes))
	}

	assertRoute(t, routes[0], http.MethodGet, "/user/browser-sessions", "browser-sessions.index", []string{"auth"})
	assertRoute(t, routes[1], http.MethodDelete, "/user/browser-sessions/{session}", "browser-sessions.destroy", []string{"auth", "password.confirm"})
	assertRoute(t, routes[2], http.MethodDelete, "/user/other-browser-sessions", "browser-sessions.destroy-other", []string{"auth", "password.confirm"})
}

func TestRoutesIncludesEnabledPasskeyContracts(t *testing.T) {
	handler := func(http.ResponseWriter, *http.Request) {}

	routes := fortify.Routes(fortify.Actions{
		BeginPasskeyRegistration:  handler,
		FinishPasskeyRegistration: handler,
		BeginPasskeyLogin:         handler,
		FinishPasskeyLogin:        handler,
	})

	if len(routes) != 4 {
		t.Fatalf("len(routes) = %d, want 4", len(routes))
	}

	assertRoute(t, routes[0], http.MethodPost, "/user/passkeys/options", "passkeys.register-options", []string{"auth", "password.confirm"})
	assertRoute(t, routes[1], http.MethodPost, "/user/passkeys", "passkeys.store", []string{"auth", "password.confirm"})
	assertRoute(t, routes[2], http.MethodPost, "/passkeys/login/options", "passkeys.login-options", []string{"guest"})
	assertRoute(t, routes[3], http.MethodPost, "/passkeys/login", "passkeys.login", []string{"guest"})
}

func TestRoutesIncludesEnabledTeamContracts(t *testing.T) {
	handler := func(http.ResponseWriter, *http.Request) {}

	routes := fortify.Routes(fortify.Actions{
		ListTeams:            handler,
		CreateTeam:           handler,
		SwitchCurrentTeam:    handler,
		AddTeamMember:        handler,
		UpdateTeamMemberRole: handler,
		RemoveTeamMember:     handler,
	})

	if len(routes) != 6 {
		t.Fatalf("len(routes) = %d, want 6", len(routes))
	}

	assertRoute(t, routes[0], http.MethodGet, "/teams", "teams.index", []string{"auth"})
	assertRoute(t, routes[1], http.MethodPost, "/teams", "teams.store", []string{"auth"})
	assertRoute(t, routes[2], http.MethodPut, "/current-team", "current-team.update", []string{"auth"})
	assertRoute(t, routes[3], http.MethodPost, "/teams/{team}/members", "team-members.store", []string{"auth"})
	assertRoute(t, routes[4], http.MethodPut, "/teams/{team}/members/{user}", "team-members.update", []string{"auth"})
	assertRoute(t, routes[5], http.MethodDelete, "/teams/{team}/members/{user}", "team-members.destroy", []string{"auth"})
}

func TestLoginHandlerAttemptsCredentials(t *testing.T) {
	guard := &stubStatefulGuard{attemptOK: true}
	handler := fortify.NewLoginHandler(guard, fortify.LoginConfig{})

	rec := perform(handler, http.MethodPost, "/login", `{"email":"taylor@example.com","password":"secret","remember":true}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if guard.attemptCredentials["email"] != "taylor@example.com" {
		t.Fatalf("attempt email = %q", guard.attemptCredentials["email"])
	}

	if guard.attemptCredentials["password"] != "secret" {
		t.Fatalf("attempt password = %q", guard.attemptCredentials["password"])
	}

	if !guard.attemptRemember {
		t.Fatal("expected remember flag to be passed to guard")
	}
}

func TestLoginHandlerReturnsValidationForInvalidCredentials(t *testing.T) {
	guard := &stubStatefulGuard{attemptOK: false}
	handler := fortify.NewLoginHandler(guard, fortify.LoginConfig{})

	rec := perform(handler, http.MethodPost, "/login", `{"email":"taylor@example.com","password":"wrong"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestLoginHandlerThrottlesFailedAttempts(t *testing.T) {
	guard := &stubStatefulGuard{attemptOK: false}
	limiter := fortify.NewMemoryLoginLimiter(2, time.Minute, time.Minute)
	handler := fortify.NewLoginHandler(guard, fortify.LoginConfig{
		Limiter:  limiter,
		LimitKey: func(*http.Request, string) string { return "login:taylor" },
	})

	rec := perform(handler, http.MethodPost, "/login", `{"email":"taylor@example.com","password":"wrong"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("first status = %d", rec.Code)
	}

	rec = perform(handler, http.MethodPost, "/login", `{"email":"taylor@example.com","password":"wrong"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("second status = %d", rec.Code)
	}

	rec = perform(handler, http.MethodPost, "/login", `{"email":"taylor@example.com","password":"wrong"}`)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("third status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}

func TestLoginHandlerClearsLimiterOnSuccess(t *testing.T) {
	guard := &stubStatefulGuard{attemptOK: false}
	limiter := fortify.NewMemoryLoginLimiter(2, time.Minute, time.Minute)
	handler := fortify.NewLoginHandler(guard, fortify.LoginConfig{
		Limiter:  limiter,
		LimitKey: func(*http.Request, string) string { return "login:taylor" },
	})

	rec := perform(handler, http.MethodPost, "/login", `{"email":"taylor@example.com","password":"wrong"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("failed status = %d", rec.Code)
	}

	guard.attemptOK = true
	rec = perform(handler, http.MethodPost, "/login", `{"email":"taylor@example.com","password":"secret"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("success status = %d", rec.Code)
	}

	guard.attemptOK = false
	rec = perform(handler, http.MethodPost, "/login", `{"email":"taylor@example.com","password":"wrong"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("post-clear status = %d", rec.Code)
	}
}

func TestRegisterHandlerCreatesUserAndCanAutoLogin(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "secret"})
	guard := &stubStatefulGuard{}

	var got fortify.RegisterInput
	handler := fortify.NewRegisterHandler(func(_ context.Context, input fortify.RegisterInput) (cauth.User, error) {
		got = input

		return user, nil
	}, guard, fortify.RegisterConfig{AutoLogin: true})

	rec := perform(handler, http.MethodPost, "/register", `{"name":"Taylor","email":"taylor@example.com","password":"secret","password_confirmation":"secret"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	if got.Email != "taylor@example.com" || got.Password != "secret" || got.Name != "Taylor" {
		t.Fatalf("unexpected registration input: %#v", got)
	}

	if guard.loginUser != user {
		t.Fatal("expected registered user to be logged in")
	}
}

func TestForgotPasswordHandlerIsEnumerationSafeAndThrottled(t *testing.T) {
	t.Run("unknown or internal error still returns ok", func(t *testing.T) {
		sender := &stubResetLinkSender{err: errors.New("user not found")}
		handler := fortify.NewForgotPasswordHandler(sender)

		rec := perform(handler, http.MethodPost, "/forgot-password", `{"email":"missing@example.com"}`)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		if sender.email != "missing@example.com" {
			t.Fatalf("email = %q", sender.email)
		}
	})

	t.Run("throttle is explicit", func(t *testing.T) {
		sender := &stubResetLinkSender{err: passwords.ErrResetLinkThrottled}
		handler := fortify.NewForgotPasswordHandler(sender)

		rec := perform(handler, http.MethodPost, "/forgot-password", `{"email":"taylor@example.com"}`)

		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
		}
	})
}

func TestResetPasswordHandlerPassesCredentialsToBroker(t *testing.T) {
	resetter := &stubPasswordResetter{}
	handler := fortify.NewResetPasswordHandler(resetter, func(_ context.Context, _ cauth.CanResetPassword, _ string, _ string) error {
		return nil
	})

	rec := perform(handler, http.MethodPost, "/reset-password", `{"email":"taylor@example.com","token":"tok","password":"secret","password_confirmation":"secret"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if resetter.credentials["email"] != "taylor@example.com" || resetter.credentials["token"] != "tok" || resetter.credentials["password"] != "secret" {
		t.Fatalf("unexpected reset credentials: %#v", resetter.credentials)
	}
}

func TestResetPasswordHandlerRejectsConfirmationMismatch(t *testing.T) {
	handler := fortify.NewResetPasswordHandler(&stubPasswordResetter{}, nil)

	rec := perform(handler, http.MethodPost, "/reset-password", `{"email":"taylor@example.com","token":"tok","password":"secret","password_confirmation":"different"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestVerificationNotificationHandlerSendsForUnverifiedUser(t *testing.T) {
	user := &verifiableUser{GenericUser: user.NewGenericUser(map[string]any{"id": "1"})}
	guard := &stubStatefulGuard{user: user}
	handler := fortify.NewVerificationNotificationHandler(guard)

	rec := perform(handler, http.MethodPost, "/email/verification-notification", `{}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !user.notified {
		t.Fatal("expected verification notification to be sent")
	}
}

func TestVerifyEmailHandlerRequiresInjectedVerifier(t *testing.T) {
	user := &verifiableUser{GenericUser: user.NewGenericUser(map[string]any{"id": "1"})}
	guard := &stubStatefulGuard{user: user}

	t.Run("missing verifier", func(t *testing.T) {
		handler := fortify.NewVerifyEmailHandler(guard, nil)

		rec := perform(handler, http.MethodPost, "/email/verify", `{}`)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})

	t.Run("verifier marks user", func(t *testing.T) {
		handler := fortify.NewVerifyEmailHandler(guard, func(_ context.Context, _ *http.Request, user cauth.MustVerifyEmail) error {
			user.MarkEmailAsVerified(time.Now())

			return nil
		})

		rec := perform(handler, http.MethodPost, "/email/verify", `{}`)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		if !user.verified {
			t.Fatal("expected email to be marked verified")
		}
	})
}

func TestConfirmPasswordHandlerStoresConfirmationTimestamp(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "hashed"})
	guard := &stubStatefulGuard{user: user}
	session := &stubConfirmationSession{values: map[string]any{}}
	handler := fortify.NewConfirmPasswordHandler(guard, &stubHasher{checkOK: true}, session)

	rec := perform(handler, http.MethodPost, "/user/confirm-password", `{"password":"secret"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if _, ok := session.values[fortify.PasswordConfirmedAtKey].(int64); !ok {
		t.Fatalf("confirmation timestamp missing: %#v", session.values)
	}
}

func TestUpdatePasswordHandlerHashesPersistsAndUpdatesUser(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "old-hash"})
	guard := &stubStatefulGuard{user: user}
	hasher := &stubHasher{checkOK: true, hash: "new-hash"}

	var persisted string
	handler := fortify.NewUpdatePasswordHandler(guard, hasher, func(_ context.Context, got cauth.User, hash string) error {
		if got != user {
			t.Fatal("unexpected user passed to updater")
		}

		persisted = hash

		return nil
	})

	rec := perform(handler, http.MethodPut, "/user/password", `{"current_password":"old","password":"new","password_confirmation":"new"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if persisted != "new-hash" {
		t.Fatalf("persisted hash = %q", persisted)
	}

	if user.GetAuthPassword() != "new-hash" {
		t.Fatalf("user password = %q", user.GetAuthPassword())
	}
}

func TestUpdatePasswordHandlerRunsSessionInvalidators(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1", "password": "old-hash"})
	guard := &stubStatefulGuard{user: user}
	called := false
	handler := fortify.NewUpdatePasswordHandler(guard, &stubHasher{checkOK: true, hash: "new-hash"}, nil, func(_ context.Context, got cauth.User) error {
		called = got == user

		return nil
	})

	rec := perform(handler, http.MethodPut, "/user/password", `{"current_password":"old","password":"new","password_confirmation":"new"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !called {
		t.Fatal("expected session invalidator to run")
	}
}

func TestAPITokenHandlersCreateListAndRevokeTokens(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1"})
	guard := &stubStatefulGuard{user: user}
	repo := patokens.NewMemoryRepository()
	issuer := patokens.NewIssuer(repo)

	create := fortify.NewCreateAPITokenHandler(guard, issuer)
	rec := perform(create, http.MethodPost, "/user/api-tokens", `{"name":"CLI","abilities":["deploy"]}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var payload struct {
		PlainText string `json:"plain_text"`
		Token     struct {
			ID        string   `json:"id"`
			Name      string   `json:"name"`
			Abilities []string `json:"abilities"`
		} `json:"token"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}

	if payload.PlainText == "" {
		t.Fatal("expected plaintext token on create")
	}

	if payload.Token.Name != "CLI" || payload.Token.Abilities[0] != "deploy" {
		t.Fatalf("unexpected token payload: %#v", payload.Token)
	}

	list := fortify.NewListAPITokensHandler(guard, repo)
	rec = perform(list, http.MethodGet, "/user/api-tokens", `{}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", rec.Code, http.StatusOK)
	}

	revoke := fortify.NewRevokeAPITokenHandler(guard, repo)
	rec = perform(revoke, http.MethodDelete, "/user/api-tokens/"+payload.Token.ID, `{"token":"`+payload.Token.ID+`"}`)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("revoke status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	if _, err := patokens.FindByPlainTextToken(context.Background(), repo, payload.PlainText); err != patokens.ErrTokenInactive {
		t.Fatalf("revoked token err = %v, want ErrTokenInactive", err)
	}
}

func TestTwoFactorHandlersEnableConfirmRegenerateAndDisable(t *testing.T) {
	user := &twoFactorUser{GenericUser: user.NewGenericUser(map[string]any{"id": "1"})}
	guard := &stubStatefulGuard{user: user}
	persistCalls := 0
	persist := func(_ context.Context, _ cauth.TwoFactorUser) error {
		persistCalls++

		return nil
	}

	enable := fortify.NewEnableTwoFactorHandler(guard, persist, fortify.TwoFactorConfig{Issuer: "Alloy"})
	rec := perform(enable, http.MethodPost, "/user/two-factor-authentication", `{}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("enable status = %d, want %d", rec.Code, http.StatusOK)
	}

	var setup struct {
		Secret        string   `json:"secret"`
		OTPAuthURL    string   `json:"otpauth_url"`
		RecoveryCodes []string `json:"recovery_codes"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&setup); err != nil {
		t.Fatal(err)
	}

	if setup.Secret == "" || setup.OTPAuthURL == "" || len(setup.RecoveryCodes) != 8 {
		t.Fatalf("unexpected setup response: %#v", setup)
	}

	if user.enabled {
		t.Fatal("two-factor should require confirmation before enabled")
	}

	code, err := twofactor.Code(setup.Secret, time.Now())

	if err != nil {
		t.Fatal(err)
	}

	confirm := fortify.NewConfirmTwoFactorHandler(guard, persist, fortify.TwoFactorConfig{})
	rec = perform(confirm, http.MethodPost, "/user/confirmed-two-factor-authentication", `{"code":"`+code+`"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want %d", rec.Code, http.StatusOK)
	}

	if !user.enabled || user.confirmedAt == nil {
		t.Fatal("expected two-factor to be enabled and confirmed")
	}

	regenerate := fortify.NewRegenerateRecoveryCodesHandler(guard, persist)
	rec = perform(regenerate, http.MethodPost, "/user/two-factor-recovery-codes", `{}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("regenerate status = %d, want %d", rec.Code, http.StatusOK)
	}

	disable := fortify.NewDisableTwoFactorHandler(guard, persist)
	rec = perform(disable, http.MethodDelete, "/user/two-factor-authentication", `{}`)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("disable status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	if user.enabled || user.secret != "" || user.confirmedAt != nil || user.recoveryCodes != nil {
		t.Fatalf("expected two-factor state to be cleared: %#v", user)
	}

	if persistCalls != 4 {
		t.Fatalf("persistCalls = %d, want 4", persistCalls)
	}
}

func TestBrowserSessionHandlersListRevokeAndRevokeOther(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1"})
	guard := &stubStatefulGuard{user: user}
	repo := browserx.NewMemoryRepository(
		browserx.Session{ID: "current", UserID: "1", IPAddress: "127.0.0.1", UserAgent: "Current", LastActiveAt: time.Now()},
		browserx.Session{ID: "other", UserID: "1", IPAddress: "127.0.0.2", UserAgent: "Other", LastActiveAt: time.Now()},
	)
	service := browserx.NewService(repo)
	current := func(*http.Request) string { return "current" }

	list := fortify.NewListBrowserSessionsHandler(guard, service, current)
	rec := perform(list, http.MethodGet, "/user/browser-sessions", `{}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload struct {
		Sessions []struct {
			ID      string `json:"id"`
			Current bool   `json:"current"`
		} `json:"sessions"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}

	if len(payload.Sessions) != 2 {
		t.Fatalf("sessions = %#v", payload.Sessions)
	}

	revoke := fortify.NewRevokeBrowserSessionHandler(guard, service)
	rec = perform(revoke, http.MethodDelete, "/user/browser-sessions/other", `{"session":"other"}`)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("revoke status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	revokeOther := fortify.NewRevokeOtherBrowserSessionsHandler(guard, service, current)
	rec = perform(revokeOther, http.MethodDelete, "/user/other-browser-sessions", `{}`)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("revoke other status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestPasskeyOptionHandlersStoreServerSideSessions(t *testing.T) {
	service, sessions := newPasskeyService(t)
	user := user.NewGenericUser(map[string]any{"id": "1"})
	guard := &stubStatefulGuard{user: user}
	key := func(*http.Request) string { return "ceremony" }

	register := fortify.NewBeginPasskeyRegistrationHandler(guard, service, key)
	rec := perform(register, http.MethodPost, "/user/passkeys/options", `{}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("registration options status = %d, want %d", rec.Code, http.StatusOK)
	}

	session, err := sessions.Get(context.Background(), "ceremony")

	if err != nil {
		t.Fatal(err)
	}

	if session.Challenge == "" {
		t.Fatal("expected registration challenge to be stored")
	}

	login := fortify.NewBeginPasskeyLoginHandler(service, key)
	rec = perform(login, http.MethodPost, "/passkeys/login/options", `{}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("login options status = %d, want %d", rec.Code, http.StatusOK)
	}

	session, err = sessions.Get(context.Background(), "ceremony")

	if err != nil {
		t.Fatal(err)
	}

	if session.Challenge == "" {
		t.Fatal("expected login challenge to be stored")
	}
}

func TestTeamHandlersCreateListSwitchAndManageMembers(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "owner"})
	guard := &stubStatefulGuard{user: user}
	service := teams.NewService(teams.NewMemoryRepository(), []teams.Role{{
		Name:        "admin",
		Permissions: []string{"members:create", "members:update", "members:delete"},
	}})

	create := fortify.NewCreateTeamHandler(guard, service)
	rec := perform(create, http.MethodPost, "/teams", `{"name":"Core"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var payload struct {
		Team struct {
			ID string `json:"id"`
		} `json:"team"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}

	list := fortify.NewListTeamsHandler(guard, service)
	rec = perform(list, http.MethodGet, "/teams", `{}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", rec.Code, http.StatusOK)
	}

	switchCurrent := fortify.NewSwitchCurrentTeamHandler(guard, service)
	rec = perform(switchCurrent, http.MethodPut, "/current-team", `{"team":"`+payload.Team.ID+`"}`)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("switch status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	addMember := fortify.NewAddTeamMemberHandler(guard, service)
	rec = perform(addMember, http.MethodPost, "/teams/"+payload.Team.ID+"/members", `{"team":"`+payload.Team.ID+`","user":"member","role":"admin"}`)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("add member status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	updateMember := fortify.NewUpdateTeamMemberRoleHandler(guard, service)
	rec = perform(updateMember, http.MethodPut, "/teams/"+payload.Team.ID+"/members/member", `{"team":"`+payload.Team.ID+`","user":"member","role":"admin"}`)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("update member status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	removeMember := fortify.NewRemoveTeamMemberHandler(guard, service)
	rec = perform(removeMember, http.MethodDelete, "/teams/"+payload.Team.ID+"/members/member", `{"team":"`+payload.Team.ID+`","user":"member"}`)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("remove member status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func assertRoute(t *testing.T, route fortify.Route, method, path, name string, middleware []string) {
	t.Helper()

	if route.Method != method || route.Path != path || route.Name != name {
		t.Fatalf("route = %s %s %s, want %s %s %s", route.Method, route.Path, route.Name, method, path, name)
	}

	if len(route.Middleware) != len(middleware) {
		t.Fatalf("middleware = %#v, want %#v", route.Middleware, middleware)
	}

	for i, want := range middleware {
		if route.Middleware[i] != want {
			t.Fatalf("middleware[%d] = %q, want %q", i, route.Middleware[i], want)
		}
	}
}

func perform(handler http.HandlerFunc, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	return rec
}

func newPasskeyService(t *testing.T) (*passkeys.Service, *passkeys.MemorySessionStore) {
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

func (g *stubStatefulGuard) User(context.Context) (cauth.User, error) { return g.user, nil }
func (g *stubStatefulGuard) Check(context.Context) bool               { return g.user != nil }
func (g *stubStatefulGuard) Guest(context.Context) bool               { return g.user == nil }
func (g *stubStatefulGuard) ID(context.Context) any {
	if g.user == nil {
		return nil
	}

	return g.user.GetAuthIdentifier()
}
func (g *stubStatefulGuard) Validate(context.Context, map[string]string) bool { return g.attemptOK }
func (g *stubStatefulGuard) Attempt(_ context.Context, credentials map[string]string, remember bool) bool {
	g.attemptCredentials = credentials
	g.attemptRemember = remember

	return g.attemptOK
}
func (g *stubStatefulGuard) Once(context.Context, map[string]string) bool { return g.attemptOK }
func (g *stubStatefulGuard) Login(_ context.Context, user cauth.User, _ bool) error {
	g.loginUser = user
	g.user = user

	return nil
}
func (g *stubStatefulGuard) LoginUsingID(context.Context, string, bool) (cauth.User, error) {
	return g.user, nil
}
func (g *stubStatefulGuard) OnceUsingID(context.Context, string) (cauth.User, error) {
	return g.user, nil
}
func (g *stubStatefulGuard) ViaRemember(context.Context) bool { return false }
func (g *stubStatefulGuard) Logout(context.Context) error {
	g.loggedOut = true
	g.user = nil

	return nil
}

func (s *stubResetLinkSender) SendResetLink(_ context.Context, email string) error {
	s.email = email

	return s.err
}

func (r *stubPasswordResetter) Reset(ctx context.Context, credentials map[string]any, resetFn passwords.ResetCallback) error {
	r.credentials = credentials

	if r.err != nil {
		return r.err
	}

	if resetFn != nil {
		return resetFn(ctx, resetUser{email: credentials["email"].(string)}, credentials["token"].(string), credentials["password"].(string))
	}

	return nil
}

func (u resetUser) GetEmailForPasswordReset() string { return u.email }

func (u *twoFactorUser) IsTwoFactorEnabled() bool            { return u.enabled }
func (u *twoFactorUser) SetTwoFactorEnabled(enabled bool)    { u.enabled = enabled }
func (u *twoFactorUser) GetTwoFactorSecret() string          { return u.secret }
func (u *twoFactorUser) SetTwoFactorSecret(secret string)    { u.secret = secret }
func (u *twoFactorUser) GetTwoFactorRecoveryCodes() []string { return u.recoveryCodes }
func (u *twoFactorUser) SetTwoFactorRecoveryCodes(codes []string) {
	u.recoveryCodes = codes
}
func (u *twoFactorUser) GetTwoFactorConfirmedAt() *time.Time { return u.confirmedAt }
func (u *twoFactorUser) SetTwoFactorConfirmedAt(at *time.Time) {
	u.confirmedAt = at
}

func (u *verifiableUser) HasVerifiedEmail() bool          { return u.verified }
func (u *verifiableUser) MarkEmailAsVerified(time.Time)   { u.verified = true }
func (u *verifiableUser) MarkEmailAsUnverified()          { u.verified = false }
func (u *verifiableUser) GetEmailForVerification() string { return "taylor@example.com" }
func (u *verifiableUser) SendEmailVerificationNotification(context.Context) {
	u.notified = true
}

func (h *stubHasher) Hash(context.Context, string) (string, error) { return h.hash, nil }
func (h *stubHasher) Check(context.Context, string, string) (bool, error) {
	return h.checkOK, nil
}
func (h *stubHasher) NeedsRehash(string) bool { return false }

func (s *stubConfirmationSession) Put(key string, value any) {
	s.values[key] = value
}

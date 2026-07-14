package fortify

import "net/http"

// Route describes a headless auth endpoint without coupling auth to a router.
type Route struct {
	Method     string
	Path       string
	Name       string
	Handler    http.HandlerFunc
	Middleware []string
}

// Actions groups the Fortify-style handlers that have been enabled by the app.
type Actions struct {
	Register                      http.HandlerFunc
	Login                         http.HandlerFunc
	Logout                        http.HandlerFunc
	ForgotPassword                http.HandlerFunc
	ResetPassword                 http.HandlerFunc
	EmailVerificationNotification http.HandlerFunc
	VerifyEmail                   http.HandlerFunc
	ConfirmPassword               http.HandlerFunc
	UpdateProfile                 http.HandlerFunc
	UpdatePassword                http.HandlerFunc
	ListAPITokens                 http.HandlerFunc
	CreateAPIToken                http.HandlerFunc
	RevokeAPIToken                http.HandlerFunc
	EnableTwoFactor               http.HandlerFunc
	ConfirmTwoFactor              http.HandlerFunc
	DisableTwoFactor              http.HandlerFunc
	RegenerateRecoveryCodes       http.HandlerFunc
	ListBrowserSessions           http.HandlerFunc
	RevokeBrowserSession          http.HandlerFunc
	RevokeOtherBrowserSessions    http.HandlerFunc
	BeginPasskeyRegistration      http.HandlerFunc
	FinishPasskeyRegistration     http.HandlerFunc
	BeginPasskeyLogin             http.HandlerFunc
	FinishPasskeyLogin            http.HandlerFunc
	ListTeams                     http.HandlerFunc
	CreateTeam                    http.HandlerFunc
	SwitchCurrentTeam             http.HandlerFunc
	AddTeamMember                 http.HandlerFunc
	UpdateTeamMemberRole          http.HandlerFunc
	RemoveTeamMember              http.HandlerFunc
}

// Routes returns the default headless Fortify route contract for enabled actions.
func Routes(actions Actions) []Route {
	routes := make([]Route, 0, 10)

	add := func(method, path, name string, handler http.HandlerFunc, middleware ...string) {
		if handler == nil {
			return
		}

		routes = append(routes, Route{
			Method:     method,
			Path:       path,
			Name:       name,
			Handler:    handler,
			Middleware: middleware,
		})
	}

	add(http.MethodPost, "/register", "register", actions.Register, "guest")
	add(http.MethodPost, "/login", "login", actions.Login, "guest")
	add(http.MethodPost, "/logout", "logout", actions.Logout, "auth")
	add(http.MethodPost, "/forgot-password", "password.email", actions.ForgotPassword, "guest")
	add(http.MethodPost, "/reset-password", "password.update", actions.ResetPassword, "guest")
	add(http.MethodPost, "/email/verification-notification", "verification.send", actions.EmailVerificationNotification, "auth", "throttle")
	add(http.MethodPost, "/email/verify", "verification.verify", actions.VerifyEmail, "auth", "signed")
	add(http.MethodPost, "/user/confirm-password", "password.confirm", actions.ConfirmPassword, "auth")
	add(http.MethodPut, "/user/profile-information", "user-profile-information.update", actions.UpdateProfile, "auth")
	add(http.MethodPut, "/user/password", "user-password.update", actions.UpdatePassword, "auth")
	add(http.MethodGet, "/user/api-tokens", "api-tokens.index", actions.ListAPITokens, "auth")
	add(http.MethodPost, "/user/api-tokens", "api-tokens.store", actions.CreateAPIToken, "auth")
	add(http.MethodDelete, "/user/api-tokens/{token}", "api-tokens.destroy", actions.RevokeAPIToken, "auth")
	add(http.MethodPost, "/user/two-factor-authentication", "two-factor.enable", actions.EnableTwoFactor, "auth", "password.confirm")
	add(http.MethodPost, "/user/confirmed-two-factor-authentication", "two-factor.confirm", actions.ConfirmTwoFactor, "auth", "password.confirm")
	add(http.MethodDelete, "/user/two-factor-authentication", "two-factor.disable", actions.DisableTwoFactor, "auth", "password.confirm")
	add(http.MethodPost, "/user/two-factor-recovery-codes", "two-factor.recovery-codes", actions.RegenerateRecoveryCodes, "auth", "password.confirm")
	add(http.MethodGet, "/user/browser-sessions", "browser-sessions.index", actions.ListBrowserSessions, "auth")
	add(http.MethodDelete, "/user/browser-sessions/{session}", "browser-sessions.destroy", actions.RevokeBrowserSession, "auth", "password.confirm")
	add(http.MethodDelete, "/user/other-browser-sessions", "browser-sessions.destroy-other", actions.RevokeOtherBrowserSessions, "auth", "password.confirm")
	add(http.MethodPost, "/user/passkeys/options", "passkeys.register-options", actions.BeginPasskeyRegistration, "auth", "password.confirm")
	add(http.MethodPost, "/user/passkeys", "passkeys.store", actions.FinishPasskeyRegistration, "auth", "password.confirm")
	add(http.MethodPost, "/passkeys/login/options", "passkeys.login-options", actions.BeginPasskeyLogin, "guest")
	add(http.MethodPost, "/passkeys/login", "passkeys.login", actions.FinishPasskeyLogin, "guest")
	add(http.MethodGet, "/teams", "teams.index", actions.ListTeams, "auth")
	add(http.MethodPost, "/teams", "teams.store", actions.CreateTeam, "auth")
	add(http.MethodPut, "/current-team", "current-team.update", actions.SwitchCurrentTeam, "auth")
	add(http.MethodPost, "/teams/{team}/members", "team-members.store", actions.AddTeamMember, "auth")
	add(http.MethodPut, "/teams/{team}/members/{user}", "team-members.update", actions.UpdateTeamMemberRole, "auth")
	add(http.MethodDelete, "/teams/{team}/members/{user}", "team-members.destroy", actions.RemoveTeamMember, "auth")

	return routes
}

// RegisterRoutes attaches route descriptors to an http.ServeMux.
func RegisterRoutes(mux *http.ServeMux, routes []Route) {
	for _, route := range routes {
		route := route
		mux.HandleFunc(route.Method+" "+route.Path, route.Handler)
	}
}

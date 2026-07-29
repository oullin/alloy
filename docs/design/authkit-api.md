# Design Spike: `authkit` / `authflows` public API

> **Status**: Design spike (plan 025). This is a design artifact, not shipped
> code. It defines a proposed public API; it does not build the package. All
> "current state" claims cite real files on `main` as of the spike branch.
> Breaking changes are acceptable for this pre-GA commercial product, so this
> doc biases toward the _right_ API over compatibility.

## 1. Summary and recommendation

The auth primitives in `pkg/hub/auth/*` are already close to feature-complete:
session guard, headless Fortify-style handlers, password reset broker, WebAuthn
passkeys, personal access tokens, browser-session management, teams, and a TOTP
two-factor implementation all exist. What is missing is a **composition layer**:
a single, container-friendly entry point that assembles those ~10 collaborators
into a working authentication stack, plus a **flow layer** that sequences the
multi-step journeys (register → verify → login, passkey enrollment, password
reset) as first-class, observable state.

The smoking-gun evidence is the reference consumer itself: `web/inertia-demo/api/auth/*`
**does not use `pkg/hub/auth` at all**. It re-implements a weaker, hand-rolled
login/session stack (`bcrypt` compare + AES-encrypted user-id cookie) because
wiring the real primitives is too much assembly for an app author. `authkit` is
the layer that makes the primitives the path of least resistance.

**Recommendation:**

1. Build `authkit` as a Go composition facade over the existing primitives
   (Section 4). Do not reimplement any primitive; compose them. One small
   primitive extension is needed (Section 7).
2. Build `authflows` as a thin, opinionated set of flow definitions on top of
   the existing `pkg/hub/workflow` Petri-net engine (Section 5), not a new
   engine.
3. Ship `authkit`/`authflows` **Go-first**. Add a **narrow** `sdk/*` TypeScript
   twin limited to the browser-side ceremony helpers (passkeys, flow-step
   client), following the existing `@hara/sdk-*` twin pattern — not a full
   port (Section 6).
4. Note `billing` (on `sdk/money`) as the obvious next composition target, but
   do not design it here (out of scope per the plan).

The rest of this document is the grounding evidence, the proposed surfaces, the
breaking-change and extension inventory, a validation sketch, and open
questions for the owner.

## 2. Current state: the primitives to compose

Every entry below is a real, constructable surface on `main`.

| Concern | Primitive | Constructor / key surface | File |
| --- | --- | --- | --- |
| Registry / DI | `auth.ServiceProvider`, `manager.Registry` | `NewServiceProvider(app, defaultGuard).WithBoot(...)`; `Registry.Extend/Provider/ViaRequest/SetConfig/Guard` | `pkg/hub/auth/provider.go`, `pkg/hub/auth/manager/manager.go` |
| Session guard | `sessionx.SessionGuard` (implements `StatefulGuard`) | `NewSessionGuard(name, provider, session, cookies, hasher)`; `Attempt/Login/Logout`, remember-me, event dispatch | `pkg/hub/auth/sessionx/session_guard.go` |
| Headless handlers | `fortify.*` | `NewLoginHandler(guard, LoginConfig)`, `NewRegisterHandler(create, guard, cfg)`, `NewCreate/List/RevokeAPITokenHandler`, verification, passkeys, 2FA, browser sessions, teams; `Actions` struct + `Routes(actions)` | `pkg/hub/auth/fortify/*.go` |
| Callback contracts | `fortify` interfaces/func types | `RegisterUser`, `VerifyEmail`, `PasswordResetter`, `LoginLimiter`, `PasskeyService`, `ProfileUpdater`, `PasswordUpdater`, … | `pkg/hub/auth/fortify/contracts.go` |
| Login throttle | `fortify.MemoryLoginLimiter` | `NewMemoryLoginLimiter(max, decay, lockout)` | `pkg/hub/auth/fortify/login_limiter.go` |
| Password reset | `passwords.Broker` | `NewBroker(users, tokens, expiry).WithThrottle(...).WithEventDispatcher(...)` | `pkg/hub/auth/passwords/broker.go` |
| API tokens | `tokens.Issuer` | `NewIssuer(repo)`; `CreateToken(ctx, user, name, abilities, expiresAt)` | `pkg/hub/auth/tokens/issuer.go` |
| Passkeys (WebAuthn) | `passkeys.Service` | `NewService(webauthn.Config, repo, sessions)`; begin/finish registration + discoverable login | `pkg/hub/auth/passkeys/service.go` |
| Browser sessions | `browserx.Service` | `NewService(repo)`; `List/Revoke/RevokeOther` | `pkg/hub/auth/browserx/repository.go` |
| Teams | `teams.Service` | `NewService(repo, roles)`; `Create/AddMember/UpdateRole/RemoveMember/SwitchCurrent` | `pkg/hub/auth/teams/service.go` |
| Two-factor | `twofactor` funcs | `GenerateSecret`, `Code`, `Verify`, `OTPAuthURL` + `recovery` codes | `pkg/hub/auth/twofactor/totp.go` |
| Authorization | `access.Gate` + middleware | `AuthorizeMiddleware(gate, ability, model)` | `pkg/hub/auth/access/gate.go`, `.../middleware.go` |
| HTTP glue | `httpx` | `EnsureAuthenticated(guard)`, `WithUser`, `UserFromContext`, `RedirectIfAuthenticated` | `pkg/hub/auth/httpx/middleware.go` |
| Session-guard migration | `sessionx` `Migrate` on login/privilege change | `SessionGuard.Login` calls `session.Migrate(ctx, true)` | `pkg/hub/auth/sessionx/session_guard.go:337` |
| Contracts | `cauth.Guard`, `StatefulGuard`, `HTTPGuard`, `User`, `UserProvider` | interface set every layer already speaks | `pkg/hub/contracts/auth/guard.go`, `.../provider.go`, `.../authenticatable.go` |
| Workflow engine | `workflow.Machine[T]`, `DefinitionBuilder` | Petri-net engine + `store`, `registry`, `events`, `multisteps` | `pkg/hub/workflow/doc.go` |

Two things stand out:

- **The pieces already share a contract vocabulary** (`cauth.User`,
  `cauth.StatefulGuard`, event dispatchers). Composition is a wiring problem,
  not an interface-mismatch problem.
- **`Routes(Actions)` already exists** (`pkg/hub/auth/fortify/routes.go:49`) and
  returns router-agnostic `Route` descriptors with `Method/Path/Name/Middleware`.
  `authkit` should _produce_ a populated `Actions`, not invent a new routing model.

## 3. Grounding evidence: what consumers hand-wire today

### 3.1 The demo bypasses the entire stack

`web/inertia-demo/api/auth` is the only in-repo consumer, and it re-implements
auth from scratch rather than composing `pkg/hub/auth`:

- **Credential check** is raw `bcrypt.CompareHashAndPassword` against a
  hand-written `FindUserByEmail` (`service.go:25-37`) — bypassing
  `SessionGuard.Attempt` and the `UserProvider` contract.
- **Sessions** are a bespoke AES-256-CBC-encrypted user-id cookie
  (`session.go:101-130`, `SessionCookieName = "inertia_go_demo_session"`) built
  directly on `encryption.NewEncrypter`. This duplicates — more weakly — what
  `SessionGuard` + the session store + remember-me token rotation already do
  (`sessionx/session_guard.go:336-370`). There is no session fixation defense
  (`Migrate`), no remember-token cycling, no password-hash-segment validation.
- **Middleware** (`WithCurrentUser`, `RequireAuth`, `GuestOnly` in
  `session.go:24-56`) is a re-implementation of `httpx.EnsureAuthenticated` /
  `RedirectIfAuthenticated`.
- **Handlers and forms** (`handlers.go`, `forms.go`) re-implement the
  login/logout HTTP surface that `fortify.NewLoginHandler` / `NewLogoutHandler`
  already provide (`fortify/login.go:21-98`).
- **Nothing else is wired at all**: no throttling, email verification, passkeys,
  API tokens, 2FA, or browser-session management — even though every one of
  those primitives exists and is tested.

The demo's host-integration shape is instructive: `Container{ DB, CryptoKey,
Render, Redirect, RouteURL, SetFlash, SecureCookie }` (`container.go:13-21`).
That is roughly the set of host hooks `authkit` will also need — but the demo
supplies them to hand-rolled logic instead of to the primitives.

### 3.2 The assembly burden, quantified

Wiring the primitives "the right way" today means an app author must, by hand:

1. Build a `UserProvider` (ORM or database) and a session store + cookie jar.
2. Construct a `SessionGuard(name, provider, session, cookies, hasher)` and set
   its event dispatcher, remember duration, and cookie attributes.
3. Register the guard in `manager.Registry` (driver/config map) via
   `ServiceProvider.WithBoot`.
4. Construct a `MemoryLoginLimiter` and a `LoginConfig`, then
   `fortify.NewLoginHandler`.
5. Construct `passwords.Broker` (users, token repo, expiry, throttle, events)
   and wire `ForgotPassword` / `ResetPassword` handlers.
6. Construct `tokens.Issuer` + repo and wire the three API-token handlers.
7. Construct `passkeys.Service(webauthn.Config, repo, sessions)` and wire four
   passkey handlers plus a `PasskeyUserResolver` and a `PasskeySessionKey`.
8. Construct `browserx.Service`, `teams.Service`, 2FA updaters, profile/password
   updaters — each with its own callback contract from
   `fortify/contracts.go`.
9. Populate the ~31-field `fortify.Actions` struct (`fortify/routes.go:15-46`)
   and call `Routes(actions)` + `RegisterRoutes(mux, routes)`.

`pkg/hub/auth/fortify/fortify_test.go` shows the collaborator surface concretely:
a single test file has to stub `StatefulGuard`, `ResetLinkSender`,
`PasswordResetter`, `PasskeyService`, hasher, confirmation session, and pull in
`browserx`, `passwords`, `teams`, `tokens`, `twofactor`, and `user` just to
exercise handlers. That stub burden _is_ the wiring burden a real app faces.

**`authkit`'s job is to collapse steps 1–9 into one configured builder.**

## 4. Proposed `authkit` API (the composition layer)

Design goals: (a) one builder that yields a working stack from config + a few
host hooks; (b) escape hatches to the underlying primitives so nothing is
hidden; (c) produce the existing `fortify.Actions`/`Routes` rather than a novel
router; (d) container/DI-friendly, mirroring `auth.ServiceProvider`.

> Signatures below are **sketches** for review, not final code. Package path is
> illustrative; the build-vs-not and final import path are owner decisions
> (Section 8). This spike does **not** create the package.

### 4.1 Configuration

```go
// Config is the declarative description of an app's auth surface. Zero values
// select safe defaults; features are opt-in so an app pays only for what it
// mounts.
type Config struct {
    // Guard identity and defaults.
    GuardName    string        // default "web"
    DefaultGuard string        // registry default; default GuardName

    // Feature toggles — each gates a slice of fortify.Actions.
    Features Features

    // Throttling for login + verification resend.
    Login LoginPolicy

    // Password reset token lifetime / throttle.
    PasswordReset PasswordResetPolicy

    // Passkey relying-party config (maps to webauthn.Config).
    Passkeys PasskeyPolicy

    // Remember-me + cookie hardening.
    Cookies CookiePolicy
}

type Features struct {
    Registration     bool
    EmailVerification bool
    PasswordReset    bool
    APITokens        bool
    Passkeys         bool
    TwoFactor        bool
    BrowserSessions  bool
    Teams            bool
}
```

### 4.2 Dependencies (host-supplied ports)

These are the collaborators `authkit` cannot invent — the app's persistence and
identity. They intentionally reuse the existing contracts.

```go
// Deps carries the app-owned adapters authkit composes. Only the fields
// required by enabled Features must be non-nil; Build validates this.
type Deps struct {
    Users        cauth.UserProvider        // required
    Sessions     sessionx.SessionStore     // required
    Cookies      sessionx.CookieManager    // required
    Hasher       cauth.PasswordHasher      // required
    Events       events.Dispatcher         // optional; enables lifecycle events

    // Feature-scoped repositories (required only when the feature is enabled).
    PasswordTokens passwords.TokenRepository
    APITokens      tokens.Repository
    Passkeys       passkeys.Repository
    PasskeySessions passkeys.SessionStore
    BrowserSessions browserx.Repository
    Teams          teams.Repository

    // Application callbacks (thin, from fortify/contracts.go).
    CreateUser     fortify.RegisterUser
    VerifyEmail    fortify.VerifyEmail
    UpdateProfile  fortify.ProfileUpdater
    UpdatePassword fortify.PasswordUpdater
    ResolvePasskeyUser fortify.PasskeyUserResolver
}
```

### 4.3 Construction and output

```go
// Kit is the assembled auth stack. It owns the guard and every enabled service,
// and exposes the router-agnostic action/route contract fortify already defines.
type Kit struct { /* unexported */ }

// New composes a Kit from config + deps. It builds the SessionGuard, registers
// it, constructs each enabled service, and populates a fortify.Actions.
// Returns an error listing every missing dependency for the enabled features.
func New(cfg Config, deps Deps) (*Kit, error)

// Guard returns the composed StatefulGuard (escape hatch to sessionx).
func (k *Kit) Guard() cauth.StatefulGuard

// Actions returns the populated fortify.Actions for enabled features only.
func (k *Kit) Actions() fortify.Actions

// Routes returns fortify.Route descriptors (Method/Path/Name/Middleware).
func (k *Kit) Routes() []fortify.Route

// Mount attaches routes to a standard mux (thin wrapper over RegisterRoutes).
func (k *Kit) Mount(mux *http.ServeMux)

// Middleware returns the auth/guest middleware pair (httpx-backed) so the app
// can protect its own routes without re-deriving them.
func (k *Kit) Middleware() (requireAuth, guestOnly func(http.Handler) http.Handler)

// Primitive accessors — deliberate escape hatches, so authkit never becomes a
// ceiling. Return nil when the feature is disabled.
func (k *Kit) Passwords() *passwords.Broker
func (k *Kit) Tokens() *tokens.Issuer
func (k *Kit) Passkeys() *passkeys.Service
func (k *Kit) BrowserSessions() *browserx.Service
func (k *Kit) Teams() *teams.Service
```

### 4.4 Container integration

Mirror the existing `auth.ServiceProvider` so `authkit` drops into the same
`container.App` lifecycle (`pkg/hub/auth/provider.go`):

```go
// ServiceProvider registers a Kit as a singleton and boots its guard into the
// auth Registry, so authkit coexists with the lower-level auth provider.
func NewServiceProvider(app *container.App, cfg Config, deps Deps) *ServiceProvider
func (p *ServiceProvider) Register()      // binds "authkit" and reuses "auth"
func (p *ServiceProvider) Boot()          // registers the guard via Registry.WithBoot
func (p *ServiceProvider) Provides() []string
```

### 4.5 What the demo would become

The ~150 lines across `web/inertia-demo/api/auth/{service,session,handlers,forms}.go`
collapse to roughly:

```go
kit, err := authkit.New(authkit.Config{
    GuardName: "web",
    Features:  authkit.Features{Registration: true, PasswordReset: true, Passkeys: true},
    Login:     authkit.LoginPolicy{MaxAttempts: 5, Decay: time.Minute, Lockout: time.Minute},
}, authkit.Deps{
    Users:    userProvider, Sessions: store, Cookies: jar, Hasher: hasher,
    CreateUser: createUser, PasswordTokens: tokenRepo,
})
if err != nil { return err }
kit.Mount(mux)
requireAuth, guestOnly := kit.Middleware()
```

That replaces bespoke bcrypt/AES/cookie/middleware code with the hardened
primitives (session fixation defense, remember-token rotation, throttling) for
free — the exact gap Section 3.1 documents.

## 5. Proposed `authflows` API (the orchestrated flows)

`authflows` sequences multi-step journeys as observable state. Rather than a
new engine, it defines flow **definitions** on top of the existing
`pkg/hub/workflow` Petri-net engine (`pkg/hub/workflow/doc.go`), reusing its
`store` (marking persistence), `events` (typed dispatch), and `audit` (trail)
subpackages. This gives resumability, auditability, and typed transition events
without new machinery.

### 5.1 The flows in scope for v1

| Flow | Places (states) | Transitions | Backing primitive |
| --- | --- | --- | --- |
| `Onboarding` | `registered → email_pending → verified → active` | `register`, `send_verification`, `verify`, `first_login` | `fortify.RegisterUser`, `VerifyEmail`, `SessionGuard.Login` |
| `PasskeyEnrollment` | `guest → challenged → enrolled` | `begin`, `finish` | `passkeys.Service` begin/finish registration |
| `PasswordReset` | `requested → link_sent → reset` | `request_link`, `reset` | `passwords.Broker.SendResetLinkUsing/Reset` |
| `PasswordlessLogin` | `guest → asserted → authenticated` | `begin_login`, `finish_login` | `passkeys.Service` discoverable login + `guard.Login` |

### 5.2 Flow contracts

```go
// Flow is a typed wrapper over a workflow.Machine plus the auth side effects
// each transition performs. T is the flow's aggregate (e.g. *Onboarding).
type Flow[T any] struct { /* unexported: machine, store, events */ }

// Step is a transition the caller drives. Apply advances the flow and runs the
// bound auth action, emitting a typed event on success.
func (f *Flow[T]) Can(ctx context.Context, subject T, step string) bool
func (f *Flow[T]) Apply(ctx context.Context, subject T, step string, input StepInput) (StepResult, error)
func (f *Flow[T]) State(subject T) string

// Builders return preconfigured flows wired to authkit's services, so an app
// does not re-declare the Petri net.
func NewOnboarding(kit *authkit.Kit, opts OnboardingOptions) *Flow[*Onboarding]
func NewPasskeyEnrollment(kit *authkit.Kit) *Flow[*Enrollment]
func NewPasswordReset(kit *authkit.Kit) *Flow[*Reset]
```

### 5.3 Events

Each transition emits a typed event through the flow's dispatcher, reusing the
`pkg/hub/auth/events` vocabulary where one already exists (`Registered`,
`Login`, `PasswordResetLinkSent` are already dispatched by the primitives —
`passwords/broker.go:169`, `sessionx/session_guard.go:366`) and adding
flow-level events (`FlowStepCompleted{Flow, Step, Subject}`,
`FlowCompleted{Flow, Subject}`) for orchestration observers. `authflows` should
**subscribe to** the primitives' existing events rather than re-emit them, to
avoid double-dispatch.

### 5.4 Why workflow-backed rather than ad-hoc

- Resumable onboarding (email verification can complete hours later) maps
  naturally to a persisted marking via `workflow/store`.
- The audit trail is a compliance asset for a commercial product.
- The `sdk/workflow` TS twin already exists (`@hara/sdk-workflow`), which makes
  the client-side flow-state mirror in Section 6 cheap.

## 6. Cross-runtime recommendation (Go-only vs. TS twin)

**Recommendation: Go-first core, narrow TS twin — not a full port.**

Rationale:

- The `sdk/*` twin pattern is established: `@hara/sdk-workflow`,
  `@hara/sdk-money`, plus `console`, `tempo`, `navigator-routes`
  (`sdk/*/package.json`). A twin is idiomatic here, not novel.
- **Passkeys genuinely need a browser client.** WebAuthn ceremonies run
  `navigator.credentials.create/get()` in the browser; the server
  (`passkeys.Service`) only issues/validates options. A thin
  `@hara/sdk-authkit` that wraps the begin/finish round-trips and the
  `PublicKeyCredential` (de)serialization removes real friction and is the
  highest-value twin surface.
- **Flow state benefits from a client mirror.** Because `sdk/workflow` already
  exists, a small `@hara/sdk-authflows` that mirrors flow step-state (which
  step am I on, what can I submit next) is cheap and improves the SPA/Inertia
  UX that the demo targets.
- **Do not** port the composition/guard/session logic to TS. Session
  management, hashing, token issuance, and DB access belong on the server; a TS
  reimplementation would fork security-critical logic across runtimes.

Concretely, the TS twin scope is:

| TS surface | Purpose | Backs onto |
| --- | --- | --- |
| `@hara/sdk-authkit` (passkeys client) | `beginRegistration()/finishRegistration()`, `beginLogin()/finishLogin()`; WebAuthn (de)serialization | `passkeys.Service` HTTP endpoints |
| `@hara/sdk-authflows` (flow-state client) | typed step-state + "can submit next" mirror | `authflows` HTTP + `sdk/workflow` |

Coordinate the final policy with plan 026 (parity policy) — flagged as an open
question (Section 8).

## 7. Primitives needing a (small) extension

Composition is mostly possible today, but a few gaps force awkward wiring.
None require breaking an existing consumer beyond the demo (see Section 9).

1. **No feature-scoped `Actions` builder.** `fortify.Actions` is a flat 31-field
   struct populated by hand (`fortify/routes.go:15-46`). `authkit` needs an
   internal helper that constructs only the handlers for enabled features. This
   is additive (a new constructor in `fortify`, or kept inside `authkit`); no
   change to `Actions` itself.
2. **`SessionGuard` construction is positional and wide.**
   `NewSessionGuard(name, provider, session, cookies, hasher)` plus a dozen
   `SetX` methods (`sessionx/session_guard.go:61-82, 646-682`). Consider an
   additive `NewSessionGuardWith(cfg SessionGuardConfig)` options constructor so
   `authkit` can pass remember-duration, cookie attributes, and dispatcher in
   one call. Additive; existing constructor stays.
3. **Passkey session-key + user-resolver are per-handler callbacks.**
   `PasskeySessionKey` and `PasskeyUserResolver` (`fortify/contracts.go:51-54`)
   are supplied to each passkey handler individually. `authkit` should set them
   once on `Config.Passkeys`. No primitive change — just centralization in the
   facade.
4. **No default `TokenRepository` throttle contract on all repos.**
   `passwords.Broker.WithThrottle` requires the repo to also implement
   `RecentTokenRepository` or it errors at runtime
   (`passwords/broker.go:135-147`). `authkit` should surface this as a
   compile-/build-time requirement in `Deps` validation rather than a runtime
   error, or document that throttle needs a `RecentTokenRepository`. Doc/validation
   change only.
5. **Event dispatcher wiring is per-primitive.** `SessionGuard`,
   `passwords.Broker`, and others each take their own dispatcher setter.
   `authkit` should fan a single `Deps.Events` into all of them. No primitive
   change; centralization only.

The important finding: **no primitive needs a breaking change to be composable.**
The extensions are additive constructors and centralized configuration.

## 8. Open questions for the owner

1. **Package placement.** `pkg/hub/authkit` (foundation module) vs. a nested
   module like `pkg/hub/auth/passkeys` uses, vs. a top-level `authkit/`. The
   passkeys precedent (nested module to keep `go-webauthn` out of the
   foundation) suggests keeping heavy deps out; `authkit` itself is light, so
   `pkg/hub/authkit` is likely fine. Owner decision.
2. **v1 feature cut.** Is v1 the "core four" (register, login+throttle, logout,
   password reset), deferring passkeys/2FA/teams/API-tokens to v1.1? Recommend
   yes — ship the highest-reuse surface first, matching the demo's actual needs.
3. **TS twin timing.** Build the `@hara/sdk-authkit` passkey client with the Go
   v1, or fast-follow? Depends on plan 026's parity policy — must be reconciled
   before committing.
4. **Teams in `authkit` vs. separate.** Teams is arguably an authorization/org
   concern, not authentication. Consider whether it belongs in `authkit` or a
   sibling `orgkit`. Flagged, not decided (org-modeling scope).
5. **Relationship to `billing`.** `billing` (on `sdk/money`) is the next
   composition target and will likely want `authkit`'s user/team identity. Do
   we define a shared identity contract now so `billing` composes cleanly later?
   Note only; `billing` is out of scope for this spike.
6. **`workflow` coupling for `authflows`.** Confirm the owner wants `authflows`
   built on `pkg/hub/workflow` (recommended) rather than a bespoke lightweight
   state enum. The workflow engine adds power (resumability, audit) but also a
   dependency; acceptable given it is a first-party package.

## 9. Breaking-change inventory

Per the commercial pre-GA posture, breaking changes are acceptable; this is the
honest accounting of what building `authkit` would touch.

| Change | Breaking? | Who is affected | Notes |
| --- | --- | --- | --- |
| New `authkit` package | No | New surface | Purely additive. |
| New `authflows` package | No | New surface | Additive; depends on `workflow`. |
| Additive `fortify` feature-scoped `Actions` builder | No | None | New constructor; `Actions`/`Routes` unchanged. |
| Additive `NewSessionGuardWith(cfg)` options constructor | No | None | Existing `NewSessionGuard` retained. |
| Rewrite `web/inertia-demo/api/auth` to use `authkit` | **Yes** (demo only) | The demo | Intended — the demo is the reference and should model best practice. Cookie name/session format would change. |
| Centralize passkey session-key/user-resolver on `Config` | No | None (new API) | Old per-handler callbacks stay available. |
| `Deps` validation upgrading throttle-repo mismatch from runtime to build/validation error | No | None | Stricter validation in new API only. |
| New TS twins `@hara/sdk-authkit`, `@hara/sdk-authflows` | No | New surface | Additive; follows existing twin pattern. |

**Net:** the only real breaking change is intentional — migrating the demo off
its hand-rolled auth onto `authkit`. No existing `pkg/hub/auth` public API needs
to break to make `authkit` possible. (This satisfies the plan's STOP condition:
the design does **not** require breaking a primitive's public API in a way that
affects consumers beyond the demo.)

## 10. Validation sketch (in lieu of a throwaway prototype)

The plan's prototype step is optional and its stop condition warns against
creating a real `sdk/authkit` or `pkg/hub/authkit` package. To honor that, the
shape is validated on paper rather than by adding a scratch package. The
"login-with-throttle-and-session" flow the plan names maps cleanly onto existing
calls:

```
authkit.New(cfg, deps)
  └─ sessionx.NewSessionGuard("web", deps.Users, deps.Sessions, deps.Cookies, deps.Hasher)
  └─ fortify.NewMemoryLoginLimiter(cfg.Login.MaxAttempts, .Decay, .Lockout)
  └─ fortify.NewLoginHandler(guard, fortify.LoginConfig{Limiter: limiter, ...})
  └─ Actions{Login: loginHandler, Logout: fortify.NewLogoutHandler(guard)}
  └─ fortify.Routes(actions) ── kit.Mount(mux)

Request → limiter.TooManyAttempts → guard.Attempt (RetrieveByCredentials →
ValidateCredentials → guard.Login → session.Migrate + remember cookie) →
limiter.Clear.
```

Every call in that chain exists today (`fortify/login.go:21-85`,
`sessionx/session_guard.go:237-370`). The only new code is the `authkit.New`
assembly and `Config→LoginConfig` mapping — which confirms `authkit` is a
composition facade, not new behavior.

**What was awkward (captured for the build plan):**

- The `fortify.Actions` struct is wide and order-independent; a feature-scoped
  builder is needed to avoid nil-handler footguns (Section 7.1).
- `MemoryLoginLimiter`'s `LimitKey` defaults to `username|host` derived from
  `RemoteAddr` (`fortify/login.go:100-112`); behind a proxy the app must supply
  a custom `LimitKey`. `authkit.Config.Login` should expose this hook.
- Remember-me cookie hardening lives across several `SessionGuard.SetX` setters;
  folding them into `CookiePolicy` removes the multi-call ceremony.

## 11. Follow-up implementation plan outline

1. **Plan A — `authkit` core (v1).** Config/Deps/Kit + `ServiceProvider`; the
   feature-scoped `Actions` builder; the additive `sessionx` options
   constructor. Migrate the demo. Ship register/login+throttle/logout/password
   reset.
2. **Plan B — `authkit` advanced features (v1.1).** Passkeys, 2FA, API tokens,
   browser sessions, teams wired through `Config.Features`.
3. **Plan C — `authflows`.** The four flows on `pkg/hub/workflow`, with events
   and audit trail.
4. **Plan D — TS twins.** `@hara/sdk-authkit` passkey client, then
   `@hara/sdk-authflows` state mirror — gated on plan 026.
5. **Later — `billing`.** Compose `sdk/money` + `authkit` identity. Separate
   spike/design; not covered here.

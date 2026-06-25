# Headless Security Contracts

This document is the frontend/backend contract for the headless auth layer.

## Fortify Routes

Routes are emitted by `fortify.Routes(actions)`. Only non-nil handlers are
registered, so applications can enable modules incrementally.

| Method | Path | Name | Middleware |
| --- | --- | --- | --- |
| POST | `/register` | `register` | `guest` |
| POST | `/login` | `login` | `guest` |
| POST | `/logout` | `logout` | `auth` |
| POST | `/forgot-password` | `password.email` | `guest` |
| POST | `/reset-password` | `password.update` | `guest` |
| POST | `/email/verification-notification` | `verification.send` | `auth`, `throttle` |
| POST | `/email/verify` | `verification.verify` | `auth`, `signed` |
| POST | `/user/confirm-password` | `password.confirm` | `auth` |
| PUT | `/user/profile-information` | `user-profile-information.update` | `auth` |
| PUT | `/user/password` | `user-password.update` | `auth` |
| GET | `/user/api-tokens` | `api-tokens.index` | `auth` |
| POST | `/user/api-tokens` | `api-tokens.store` | `auth` |
| DELETE | `/user/api-tokens/{token}` | `api-tokens.destroy` | `auth` |
| POST | `/user/two-factor-authentication` | `two-factor.enable` | `auth`, `password.confirm` |
| POST | `/user/confirmed-two-factor-authentication` | `two-factor.confirm` | `auth`, `password.confirm` |
| DELETE | `/user/two-factor-authentication` | `two-factor.disable` | `auth`, `password.confirm` |
| POST | `/user/two-factor-recovery-codes` | `two-factor.recovery-codes` | `auth`, `password.confirm` |
| GET | `/user/browser-sessions` | `browser-sessions.index` | `auth` |
| DELETE | `/user/browser-sessions/{session}` | `browser-sessions.destroy` | `auth`, `password.confirm` |
| DELETE | `/user/other-browser-sessions` | `browser-sessions.destroy-other` | `auth`, `password.confirm` |
| POST | `/user/passkeys/options` | `passkeys.register-options` | `auth`, `password.confirm` |
| POST | `/user/passkeys` | `passkeys.store` | `auth`, `password.confirm` |
| POST | `/passkeys/login/options` | `passkeys.login-options` | `guest` |
| POST | `/passkeys/login` | `passkeys.login` | `guest` |
| GET | `/teams` | `teams.index` | `auth` |
| POST | `/teams` | `teams.store` | `auth` |
| PUT | `/current-team` | `current-team.update` | `auth` |
| POST | `/teams/{team}/members` | `team-members.store` | `auth` |
| PUT | `/teams/{team}/members/{user}` | `team-members.update` | `auth` |
| DELETE | `/teams/{team}/members/{user}` | `team-members.destroy` | `auth` |

## JSON Errors

Headless controllers should render errors through `httpx.ExceptionRenderer`
or return the same shape:

```json
{
  "message": "validation failed",
  "status": 422,
  "errors": {
    "email": ["required"]
  }
}
```

Unexpected errors render as HTTP 500 without leaking the internal error string
unless debug rendering is explicitly enabled.

## Production Defaults

Start production auth configuration from `security.ProductionDefaults()` and
then set deployment-specific values:

```go
cfg := security.ProductionDefaults()
cfg.AppKey = appKey
cfg.Passkeys.RPID = "example.com"
cfg.Passkeys.RPDisplayName = "Alloy"
cfg.Passkeys.RPOrigins = []string{"https://example.com"}

if err := cfg.ValidateProduction(); err != nil {
    return err
}
```

Production validation requires:

- app key configured
- secure and HTTP-only cookies
- CSRF origin verification
- positive session and password reset lifetimes
- login throttling and lockout
- WebAuthn relying party ID and origins

## Request Context And Logging

Use `observability.Middleware` to attach stable request metadata to the request
context and emit structured logs through the auth logger contract.

The propagated context contains:

- request ID
- authenticated user ID when available
- IP address
- user agent
- method
- path

Handlers can retrieve it with:

```go
meta, ok := observability.RequestContextFromContext(r.Context())
```

## Cache And Rate Limits

Use `github.com/oullin/alloy/cache` for shared TTL cache and fixed-window
rate-limit primitives. The memory store is suitable for tests and single-process
apps; production multi-instance deployments should provide a distributed
implementation of `cache.Store`.

Auth's `fortify.MemoryLoginLimiter` remains available for local login
throttling, while router-level throttling can use the existing routing
middleware rate-limiter contract.

## CSRF

Unsafe browser requests should use `session.VerifyCSRFToken`.

Accepted token sources:

- `X-CSRF-Token`
- `X-XSRF-Token`
- Form field `_token`

The middleware intentionally does not accept the `XSRF-TOKEN` cookie alone,
because browsers attach cookies automatically.

## Password Reset

Password reset tokens are plaintext only at delivery time. Repositories store
SHA-256 token hashes and compare submitted tokens in constant time.

The forgot-password endpoint returns the same successful response even when the
email is unknown, except explicit throttle errors return HTTP 429.

## API Tokens

Personal access tokens are returned once as `id|secret`. Storage only keeps the
secret hash.

Create response:

```json
{
  "token": {
    "id": "1",
    "name": "CLI",
    "abilities": ["deploy"],
    "created_at": "2026-06-25T00:00:00Z"
  },
  "plain_text": "1|secret"
}
```

Bearer authentication expects:

```http
Authorization: Bearer 1|secret
```

Use `tokens.RequireAbility("ability")` to enforce token abilities.

## Two Factor

Enabling 2FA generates a TOTP secret and plaintext recovery codes. Store only
hashed recovery codes on the user record. Confirmation is required before
`TwoFactorAuthenticatable.IsTwoFactorEnabled()` should become true.

## Password And Session Invalidation

`fortify.NewUpdatePasswordHandler` accepts optional
`PasswordSessionInvalidator` callbacks. Use these callbacks to revoke other
browser sessions, personal access tokens, or other credentials after a password
change.

## Browser Sessions

Browser-session listing reads the database session table metadata:

- session ID
- IP address
- user agent
- last activity timestamp
- whether it is the current session

Revocation is user scoped. A user cannot revoke another user's session by ID.

## Passkeys

The passkey layer uses `github.com/go-webauthn/webauthn`.

The frontend flow is:

1. Call `POST /user/passkeys/options` or `POST /passkeys/login/options`.
2. Pass the returned options to the browser WebAuthn API.
3. Post the browser credential response to `/user/passkeys` or
   `/passkeys/login`.

`webauthn.SessionData` must stay server-side in `passkeys.SessionStore` between
the options and finish calls. Do not expose or trust it from the browser.

## Teams

The teams module is headless and role-based. `teams.Service` supports:

- create team
- list teams for a user
- switch current team
- add member
- update member role
- remove member

Owners can manage members. Other roles are configured with named permissions
such as `members:create`, `members:update`, and `members:delete`.

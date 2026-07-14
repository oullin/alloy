# Plan 025 (SPIKE): Define the `authkit`/`authflows` public API composing the existing auth primitives

> **Executor instructions**: This is a **design/investigation spike**, not a build task. Produce a design document (and optional throwaway prototype) defining the public API; do not build the package. Deliverable: `docs/design/authkit-api.md`. Update `plans/README.md` when done.

## Status

- **Priority**: P3
- **Effort**: L (spike-scoped)
- **Risk**: MED (new public API surface pre-1.0 — breaking changes acceptable per stated goal)
- **Depends on**: none (informed by the auth correctness/security plans, but doesn't require them)
- **Category**: direction
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

`web/docs/getting-started.md:129-143` lists 24 packages "documented as design targets but not yet available", including the commercially-loaded `authkit`, `authflows`, and `billing`. These are the packages a subscription product most needs, and they're disproportionately cheap here because the primitives already exist: `pkg/hub/auth` (access gates, browserx sessions, fortify login/tokens/email-verification, the nested `passkeys` module using `go-webauthn`), `pkg/hub/session`, `hashing`, `encryption`. `authkit`/`authflows` sit one layer above already-built pieces. Defining their public API turns the roadmap into revenue-relevant surface. This spike scopes `authkit`/`authflows` first (highest reuse); `billing` (on `money`) is noted but not designed here.

## Current state (primitives to compose)

- `pkg/hub/auth/access` — gates/middleware.
- `pkg/hub/auth/browserx` — browser sessions (repo, revocation).
- `pkg/hub/auth/fortify` — login, api_tokens, email_verification, login_limiter, passkeys wiring.
- `pkg/hub/auth/passkeys` (nested module) — WebAuthn.
- `pkg/hub/auth/sessionx` — session guard (login/privilege-change `Migrate`).
- `pkg/hub/session`, `pkg/hub/hashing`, `pkg/hub/encryption`, `pkg/hub/cookie`.
- The demo API (`web/inertia-demo/api/auth`) hand-wires several of these — a signal of what `authkit` should absorb. Read it to see the friction consumers face today.

## Scope

**In scope**: a design doc defining the `authkit`/`authflows` public API (types, entry points, how they compose the primitives), a parity/scope note vs. the Laravel Fortify/Sanctum model the code echoes, and open questions. An optional throwaway prototype (a single flow wired end-to-end in a scratch file) to validate the shape — not a real package.

**Out of scope**: building the package; `billing` design (note it, don't design it); the private-distribution question (plan 024).

## Investigation steps

### Step 1: Inventory what consumers hand-wire today

Read `web/inertia-demo/api/auth/*` and any other consumer wiring. List the boilerplate (session setup, login + throttle, email verification, passkey registration/login, token issuance) that `authkit` would encapsulate. This is the grounding evidence for the API.

### Step 2: Define `authkit` (the composition layer)

Specify the public surface: how a consumer configures and constructs an authkit instance from the primitives (DI/container-friendly), the high-level operations (register, login w/ throttle, logout, verify email, issue/revoke API token, passkey register/authenticate), and how it plugs into an HTTP stack (middleware). Prefer composing the existing primitives over reimplementing them; call out any primitive that needs a small extension to be composable.

### Step 3: Define `authflows` (the orchestrated flows)

Specify the multi-step flows (e.g. registration → email verification → first login; passwordless/passkey enrollment; password reset using `auth/passwords`) as composable state, ideally reusing `pkg/hub/workflow` where it fits. Define the flow contracts and events.

### Step 4: Cross-runtime question

Decide whether `authkit`/`authflows` are Go-only or also get a `sdk/*` TS twin (relates to plan 026's parity policy). Document the recommendation with rationale (the frontend likely needs client helpers for passkeys/flows).

### Step 5: Optional prototype

In a throwaway location (not `sdk/`, not a new published package — e.g. a scratch `spike/` dir or a `_test.go` that's deleted after), wire one flow (login-with-throttle-and-session) using the real primitives to validate the proposed API shape. Capture what was awkward.

### Step 6: Write the design doc

`docs/design/authkit-api.md`: the consumer-boilerplate evidence, the `authkit` API, the `authflows` API, the Go-only-vs-twin recommendation, primitives needing extension, open questions (what's in scope for v1, how it relates to `billing`), and a follow-up implementation plan outline.

## Deliverable / done criteria

- [ ] `docs/design/authkit-api.md` exists with: consumer-friction inventory (grounded in the demo wiring), proposed `authkit` public API, proposed `authflows` API, cross-runtime recommendation, list of primitives needing extension, and open questions.
- [ ] Any prototype code is in a throwaway location and does not add a new published package or `sdk/*` entry.
- [ ] `billing` is noted as follow-up, not designed.
- [ ] `plans/README.md` row for 025 updated.

## STOP conditions

- The design would require breaking an existing auth primitive's public API in a way that affects current consumers beyond the demo — surface it as an open question, don't decide it in the spike.
- A prototype tempts you to create a real `sdk/authkit` or `pkg/hub/authkit` package — STOP; this spike defines the API, it does not build the package.

## Maintenance notes

- This spike's API doc should spawn the actual build plan(s); keep it a design artifact.
- Coordinate the cross-runtime decision with plan 026 (parity policy).

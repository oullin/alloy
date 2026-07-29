# Plan 013: Remove the committed demo auth key and known-credential seed accounts

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- web/inertia-demo/api`
> On change, reconcile excerpts; on mismatch, STOP.
>
> **Security note**: This plan never writes any key material into a file. It references only the file path and credential type. The committed key must be treated as **burned** and rotated, not merely deleted.

## Status

- **Priority**: P1
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

The demo API's session cookie is `EncryptString(userID)` under an AES key that is **committed to the repo** at `web/inertia-demo/api/cmd/resources/crypto.yml` (a git-tracked, low-entropy key). The same key backs CSRF and the flash store. Because encrypt-then-MAC uses that same public key, anyone with repo access can mint a valid session cookie for any user ID — a complete authentication bypass on any deployed instance of the demo. Compounding it, four seeded accounts share the literal password `"password"`, and the login path leaks user existence via a bcrypt short-circuit timing difference. The demo is the README's showcase for the packages; shipping it with a public master key undermines the security story of the whole product.

## Current state

- `web/inertia-demo/api/cmd/resources/crypto.yml:1` — a git-tracked base64 AES-256 key (confirmed committed via `git ls-files`). **Do not print its value.**
- `web/inertia-demo/api/auth/session.go:101-127` — session cookie = `EncryptString(userID)`; `loadCurrentUser` (67-99) trusts whatever user ID decrypts. Same key wired into CSRF (`cmd/main.go:90`) and flash (`cmd/main.go:115`).
- The crypto config loader supports env overrides (the audit noted `INERTIA_CRYPTO_*`-style overrides exist — confirm the exact variable names in `cmd/cryptoconfig.go`).
- `web/inertia-demo/api/internal/seed/users.go:19-29` — four users seeded with password `"password"`.
- `web/inertia-demo/api/auth/service.go:32` — `if user == nil || bcrypt.CompareHashAndPassword(...) != nil` — bcrypt is skipped when the user doesn't exist (timing enumeration).

Convention: config is loaded via viper with env overrides; secrets belong in the environment, not committed files. The repo has no `.env.example` yet (see plan 023 for the broader env-doc work — this plan adds the crypto/demo entries).

## Commands you will need

| Purpose                       | Command                                                      | Expected        |
| ----------------------------- | ------------------------------------------------------------ | --------------- |
| Demo API build                | `cd web/inertia-demo/api && go build ./...`                  | exit 0          |
| Demo API tests                | `cd web/inertia-demo/api && go test ./...`                   | exit 0          |
| Confirm key no longer tracked | `git ls-files web/inertia-demo/api/cmd/resources/crypto.yml` | empty (removed) |
| Format                        | `pnpm exec vp run format`                                    | exit 0          |

## Scope

**In scope**: `web/inertia-demo/api/cmd/cryptoconfig.go` (load key from env), `crypto.yml` (remove the secret; keep only non-secret structure or delete), a `web/inertia-demo/api/.env.example` entry for the crypto key, `internal/seed/users.go` (seed passwords), `auth/service.go` (login timing), and the relevant tests.

**Out of scope**: the encryption package itself; framework cookie defaults (plan 014); the broader `.env.example` for all demo vars (plan 023, though adding the crypto entry here is fine).

## Git workflow

- Branch: `advisor/013-demo-auth-security`; commit per concern; conventional-commit style.

## Steps

### Step 1: Load the crypto key from the environment; remove it from the repo

Change the crypto config loader (`cmd/cryptoconfig.go`) to require the key from an environment variable (e.g. `APP_CRYPTO_KEY` — use the existing override name if one is defined). On a missing key, fail fast at startup with a clear message (do not fall back to a built-in default). Remove the key value from `crypto.yml` (delete the file if it now holds only the secret, or blank the field). Add a placeholder line to `web/inertia-demo/api/.env.example` documenting the variable and how to generate a value (describe the command; do not embed a real key). Add a note that the previously committed key is compromised and any environment that used it must rotate.

**Verify**: `git ls-files web/inertia-demo/api/cmd/resources/crypto.yml` is empty (or the file contains no key material); `cd web/inertia-demo/api && go build ./...` → exit 0; starting without the env var fails fast (a small test or manual check).

### Step 2: Stop seeding known passwords

In `internal/seed/users.go`, replace the shared `"password"` with either a per-seed random password printed to stdout on first run, or gate seeding behind an explicit dev-only flag/env so a deployed demo doesn't ship known logins.

**Verify**: `grep -n '"password"' web/inertia-demo/api/internal/seed/users.go` returns nothing; `cd web/inertia-demo/api && go test ./...` → exit 0.

### Step 3: Make login timing-equal

In `auth/service.go:32`, when the user is not found, compare the supplied password against a fixed dummy bcrypt hash so both branches perform equivalent work, and return an identical generic error for "no such user" and "wrong password".

**Verify**: `go test ./auth -run Login` → both missing-user and wrong-password return the same generic error; (optional) a coarse timing assertion or a code-path assertion that bcrypt runs in both branches.

### Step 4: Format

**Verify**: `pnpm exec vp run format` → exit 0.

## Test plan

- `auth` tests: login returns a generic error for both missing user and wrong password; a valid credential still authenticates.
- Startup: missing crypto env var fails fast (test or documented manual check).
- Seed: no literal shared password remains.

## Done criteria

- [ ] `git ls-files web/inertia-demo/api/cmd/resources/crypto.yml` shows no committed key material.
- [ ] The crypto key is loaded from an env var; missing → fail fast.
- [ ] `.env.example` documents the key variable (placeholder only, no real value) and notes rotation of the burned key.
- [ ] No `"password"` literal in the seed; login is timing-equal.
- [ ] `cd web/inertia-demo/api && go build ./... && go test ./...` exit 0; no out-of-scope files modified; `plans/README.md` row for 013 updated.

## STOP conditions

- Removing `crypto.yml` breaks a loader that unconditionally reads the file — adjust the loader in the same plan, but if the file is referenced by unrelated tooling, report first.
- No env-override path exists in the loader and adding one touches shared config code beyond the demo — report.
- The `crypto.yml` key value would need to appear anywhere in your changes — it must not; STOP and reconsider the approach.

## Maintenance notes

- The removed key is compromised permanently; the maintainer must rotate it in any environment where it was ever deployed and invalidate cookies signed with it. State this in the PR description (not in a committed file with the old value).
- Reviewer must confirm no key material appears in the diff, `.env.example`, or commit messages.

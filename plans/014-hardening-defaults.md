# Plan 014: Secure-by-default framework primitives (Argon2 cost, cookie Secure, request size cap)

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/hashing pkg/hub/contracts/cookie pkg/hub/session pkg/hub/httpx/middleware`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P1
- **Effort**: S
- **Risk**: LOW-MED
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

Three framework defaults lean insecure, so a consumer who trusts the defaults gets a weak posture:
1. **S2**: Argon2 default memory is 1 MiB (`argonDefaultMemory = 1024`), ~19× below the RFC 9106/OWASP minimum (≥19 MiB; Laravel uses 64 MiB). Password hashes computed with the shipped defaults are cheap to crack on GPU/ASIC.
2. **S3**: Cookie/session `Secure` defaults to false framework-wide; an operator who deploys without explicitly enabling it emits session/auth cookies that can travel over plaintext HTTP.
3. **S4**: The post-size middleware only checks the `Content-Length` header; for `Transfer-Encoding: chunked` (where `ContentLength == -1` and the header is stripped) it never wraps `r.Body`, so an unbounded chunked body streams straight past the cap.

## Current state

- `pkg/hub/hashing/argon.go:62-64`: `argonDefaultMemory uint32 = 1024`, `argonDefaultTime uint32 = 2`, `argonDefaultThreads uint8 = 2`; used by `NewArgonHasher` (71-79) and `NewArgon2IdHasher` (`argon2id.go:16-33`). `NeedsRehash` (~149) already supports transparent upgrade. (bcrypt at `bcrypt.go:17` cost 12 is fine — leave it.)
- `pkg/hub/contracts/cookie/cookie.go`: `Secure *bool` (13); `DefaultOptions()` sets HTTPOnly/SameSite=Lax but leaves `Secure` nil→false; merge only turns it on (49).
- `pkg/hub/session/middleware.go`: `Secure bool` (15-16) defaults false; `mergeConfig` can only enable (55-56); cookie emitted with `Secure: w.cfg.Secure` (98).
- `pkg/hub/httpx/middleware/validate_post_size.go:21-35`: checks `r.ContentLength > maxBytes` and the `Content-Length` header string; never wraps `r.Body`.
  ```go
  if r.ContentLength > m.maxBytes { http.Error(...); return }
  if cl := r.Header.Get("Content-Length"); cl != "" { if size, err := ...; size > m.maxBytes { http.Error(...); return } }
  next.ServeHTTP(w, r)
  ```

Convention: `http.MaxBytesReader(w, r.Body, n)` is the standard stream cap. Secure-by-default with a documented local-dev opt-out is the target posture.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Go tests (touched) | `cd pkg/hub && go test ./hashing/... ./session/... ./httpx/... ./contracts/...` | exit 0 |
| Full Go suite | `pnpm exec vp run go:test` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `pkg/hub/hashing/argon.go` (+`argon2id.go`), `pkg/hub/contracts/cookie/cookie.go`, `pkg/hub/session/middleware.go`, `pkg/hub/httpx/middleware/validate_post_size.go`, and their tests.

**Out of scope**: bcrypt; the encryption package; the demo's own cookie wiring (plan 013).

## Git workflow

- Branch: `advisor/014-hardening-defaults`; commit per concern; conventional-commit style.

## Steps

### Step 1 (S2): Raise Argon2 defaults and lock a floor with a test

Raise `argonDefaultMemory` to a modern minimum (recommend 19–64 MiB; pick one and document the choice against RFC 9106/OWASP). Keep `time`/`threads` per OWASP guidance. Add a test asserting the shipped defaults meet a minimum floor so they can't silently regress. `NeedsRehash` already upgrades old hashes transparently — confirm it triggers for hashes made with the old params.

**Verify**: `cd pkg/hub && go test ./hashing -run 'Argon|Default'` → defaults meet the floor; a hash made with old params reports `NeedsRehash == true`.

### Step 2 (S3): Secure cookies by default with a documented opt-out

Change the default so `Secure` is true (or auto-enabled outside local/dev), and require an explicit opt-out for local HTTP. In `contracts/cookie/cookie.go` `DefaultOptions()` and `session/middleware.go` config defaulting, flip the default and ensure the opt-out path (e.g. a `Secure=false` override for dev) still works. Document the toggle.

**Verify**: `go test ./contracts/cookie ./session -run Secure` → default options have `Secure` true; an explicit dev override can still disable it.

### Step 3 (S4): Cap the request body stream, not just the header

In `validate_post_size.go`, in addition to the fast-path header rejection, set `r.Body = http.MaxBytesReader(w, r.Body, m.maxBytes)` so oversize chunked/undeclared bodies fail on read. Keep the early header rejection for the declared-oversize fast path.

**Verify**: `go test ./httpx/middleware -run PostSize` → a chunked request (no Content-Length) whose body exceeds the cap is rejected when a handler reads the body; a within-limit request passes.

### Step 4: Full suite + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

- `hashing`: default params meet the floor; `NeedsRehash` upgrades legacy hashes.
- `cookie`/`session`: `Secure` true by default; dev opt-out works.
- `httpx/middleware`: oversize chunked body rejected on read; the `MaxBytesReader` limit is enforced.

## Done criteria

- [ ] `grep -n "argonDefaultMemory" pkg/hub/hashing/argon.go` shows a value ≥ 19×1024; a floor test exists.
- [ ] Cookie/session `Secure` defaults to true (test asserts); documented dev opt-out.
- [ ] `grep -n "MaxBytesReader" pkg/hub/httpx/middleware/validate_post_size.go` matches; chunked-body test passes.
- [ ] `pnpm exec vp run go:test` exits 0; no out-of-scope files modified; `plans/README.md` row for 014 updated.

## STOP conditions

- Raising Argon2 memory makes the existing hashing tests exceed CI memory/time budgets — tune to the OWASP minimum (19 MiB) rather than 64, and report.
- Flipping `Secure` to default-true breaks the demo/local dev flow with no clean opt-out — coordinate with plan 013's `APP_SECURE_COOKIES` handling and report.
- Excerpts don't match live code (drift); a verification fails twice.

## Maintenance notes

- Document the Argon2 params and the secure-cookie default in the package docs so consumers understand the posture and the dev opt-out.
- Reviewer should confirm the `MaxBytesReader` limit matches (or is stricter than) any per-handler body limits already present (e.g. Inertia's 32 MiB `MaxBytesReader` in `inertia/protocol/form.go`).

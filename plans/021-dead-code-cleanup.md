# Plan 021: Remove the dead `cache.RateLimiter` abstraction and the empty `packages/` directory

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/cache packages`
> On change, reconcile; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: tech-debt
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

- **X6**: `cache.RateLimiter` (`Hit`/`TooManyAttempts`/`Clear`) has **no production consumer** — its only reference is `cache/memory_store_test.go`. Meanwhile auth login throttling uses `auth/fortify/login_limiter.go`'s `MemoryLoginLimiter`, and HTTP throttling uses a different `RateLimiter` contract via `httpx/routing/middleware/throttle_requests.go`. Three parallel limiter contracts, one of them dead, in a core package — readers assume `cache.RateLimiter` is the throttling primitive when nothing uses it, and a future consumer may wire the wrong one.
- The empty `packages/` directory is a leftover from the `packages/foundation → pkg/hub` rename (commit `bfface5`); nothing references `packages/foundation` anymore. It can mislead tooling/agents into thinking a `packages/` workspace exists.

## Current state

- `pkg/hub/cache/rate_limiter.go` — defines `cache.RateLimiter`; grep confirms the only non-test reference is `pkg/hub/cache/memory_store_test.go:58`.
- `auth/fortify/login_limiter.go` (`MemoryLoginLimiter`, used at `login.go:60`) and `httpx/routing/middleware/throttle_requests.go:16` (`= cmiddleware.RateLimiter`) are the two *live* limiter contracts.
- `packages/` — empty (`ls -la packages` shows only `.`/`..`); no source/config/doc references `packages/foundation` (grep across `.go`/`.ts`/`.json`/`.md`/`.yml` returns nothing). Note: `pnpm-workspace.yaml` lists `sdk/*` etc., not `packages/*` — but double-check it doesn't glob `packages/` before deleting.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Confirm no consumer | `grep -rn "cache.RateLimiter\|RateLimiter" pkg/hub/cache && grep -rln "cache\"" pkg/hub \| xargs grep -l RateLimiter` | only test refs |
| Cache tests | `cd pkg/hub && go test ./cache/...` | exit 0 |
| Full Go suite | `pnpm exec vp run go:test` | exit 0 |
| Workspace still resolves | `pnpm install --frozen-lockfile` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `pkg/hub/cache/rate_limiter.go` (+ its test references), the empty `packages/` directory, and a possible `pnpm-workspace.yaml` check.

**Out of scope**: the two live limiters (`fortify` and `httpx` throttle) — do not touch them; consolidating the three contracts onto one is a larger design change explicitly deferred.

## Git workflow

- Branch: `advisor/021-dead-code-cleanup`; commit per concern; conventional-commit style.

## Steps

### Step 1: Confirm `cache.RateLimiter` is truly unused

Re-verify no production code (in `pkg/hub`, the demo API, or the SDKs) imports/uses `cache.RateLimiter`:
`grep -rn "RateLimiter" pkg/hub web/inertia-demo | grep -v _test.go`. The only hits should be in `pkg/hub/cache` itself.

**If any production consumer exists, STOP** — this becomes a consolidation decision, not a deletion.

### Step 2: Remove `cache.RateLimiter`

Delete `rate_limiter.go` and remove its now-dead references from `cache/memory_store_test.go` (delete the test that only exercised the dead type, or repoint it if it also covers live `MemoryStore` behavior — read it first to avoid dropping real coverage).

**Verify**: `cd pkg/hub && go build ./... && go test ./cache/...` → exit 0.

### Step 3: Remove the empty `packages/` directory

Confirm empty and unreferenced (`find packages -type f` returns nothing; `grep -rn "packages/foundation\|/packages/" --include=*.go --include=*.ts --include=*.json --include=*.yml --include=*.md .` returns nothing relevant; `pnpm-workspace.yaml` does not glob `packages/`). Then remove it (`git rm -r packages` — since it's empty in git, it may not be tracked; if untracked, just `rmdir`).

**Verify**: `pnpm install --frozen-lockfile` → exit 0 (workspace still resolves); `ls packages` → not present.

### Step 4: Full suite + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

- No new tests. Confirm the cache package still builds and its remaining (live `MemoryStore`) tests pass. Confirm the pnpm workspace resolves after removing `packages/`.

## Done criteria

- [ ] `grep -rn "RateLimiter" pkg/hub/cache` returns nothing (type removed).
- [ ] `pkg/hub/cache` builds and tests pass; no live `MemoryStore` coverage was lost.
- [ ] `packages/` no longer exists; `pnpm install --frozen-lockfile` succeeds.
- [ ] No out-of-scope files modified; `plans/README.md` row for 021 updated.

## STOP conditions

- Any production consumer of `cache.RateLimiter` is found — STOP; this is a consolidation decision, not a deletion.
- `pnpm-workspace.yaml` (or another tool) actually references `packages/` — STOP and reconcile before deleting.
- Deleting the dead test would drop coverage of live `MemoryStore` behavior mixed into the same test — split it instead of deleting wholesale.

## Maintenance notes

- The remaining two limiter contracts (`fortify`, `httpx`) are a known duplication (deferred consolidation); note that removing the dead third one narrows the confusion but doesn't resolve it.
- Reviewer should confirm no live coverage was lost when trimming the cache test.

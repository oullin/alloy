# Plan 017: Skip unmodified session writes and move GC off the request path

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/session`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: MED
- **Depends on**: none (coordinate with plan 010 which also touches session; land 010 first if both are scheduled)
- **Category**: perf
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

- **P5**: `Save` serializes the attributes and writes to the store on **every** request — including read-only requests that never touch the session. With `DatabaseHandler` that's an INSERT/UPDATE per request; with `FileHandler` a file write per request. This is per-request write amplification against the session backend.
- **P6**: Session GC runs **synchronously inside the request goroutine** on ~`GCProbability`% of requests — a directory walk (`FileHandler`) or an unindexed bulk `DELETE` (`DatabaseHandler`) lands in that request's latency budget, inflating p99 for the unlucky users.

## Current state

- `pkg/hub/session/middleware.go`:
  - `store.Save(r.Context())` on every request (line 142).
  - Probabilistic GC inline (151-153): `if cfg.GCProbability > 0 && rand.IntN(100) < cfg.GCProbability { _ = handler.GC(r.Context(), cfg.GCMaxLifetime) }`.
- `pkg/hub/session/store.go:86-97` — `Save` unconditionally `serialize(s.attributes)` + `handler.Write(...)` with no modified/dirty check.
- Handlers: `handlers/file.go:97-129` (ReadDir + per-file Info), `handlers/database.go:108-112` (bulk DELETE), `handlers/array.go:85` (whole-map sweep under Lock).

Convention: track mutation on the attribute bag; a sliding-expiry `last_activity` refresh may still need a lightweight write on an interval even when attributes are unchanged — preserve that behavior explicitly, don't skip it silently.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Session tests | `cd pkg/hub && go test ./session/...` | exit 0 |
| Race | `cd pkg/hub && go test -race ./session/...` | exit 0 |
| Full Go suite | `pnpm exec vp run go:test` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `pkg/hub/session/store.go` (dirty tracking), `pkg/hub/session/middleware.go` (conditional save + background GC), and the session handlers/tests as needed for the GC scheduling.

**Out of scope**: the silent-drop `EncryptedStore.Put` fix (plan 010); session security defaults (plan 014); the serialization format.

## Git workflow

- Branch: `advisor/017-session-perf`; commit per concern; conventional-commit style.

## Steps

### Step 1: Track a dirty flag

Add a `dirty bool` (or a modification counter) to the session, set on any attribute mutation (`Put`/`Forget`/`Flush`/`Regenerate`/`Migrate`/token operations — find all mutators via `grep -n "s.attributes" pkg/hub/session/store.go`). Reset it after a successful `Save`.

**Verify**: `go test ./session -run Dirty` → a fresh read-only session reports not-dirty; any mutation sets dirty.

### Step 2: Skip the write when clean (but preserve sliding expiry)

In `Save` (or the middleware call at line 142), skip `handler.Write` when not dirty. If the session uses sliding-expiry `last_activity`, still perform a lightweight touch — but only on an interval (e.g. once per N seconds), not every request. Make the touch behavior explicit and configurable; document it.

**Verify**: `go test ./session -run Save` → a read-only request through a fake handler issues no `Write` (or only an interval touch); a mutating request writes.

### Step 3: Move GC to a background goroutine

Replace the inline `handler.GC(...)` with a background sweep: a single-flight/ticker-driven goroutine (only one sweep at a time), decoupled from the request path. The middleware may *trigger* a sweep probabilistically but must not block the request on it. Ensure the goroutine is tied to a lifecycle context so it stops cleanly (no leak).

**Verify**: `go test -race ./session -run GC` → GC runs off the request goroutine, only one concurrent sweep, and stops on context cancel (no leaked goroutine — use `goleak` if available, else a deterministic shutdown assertion).

### Step 4: Full suite + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

- `store_test.go`: dirty tracking; clean session doesn't write; mutating session writes.
- `middleware_test.go`: read-only request issues no store write (fake handler counts writes); GC runs in the background and doesn't block the request; background sweep is single-flighted and shuts down cleanly.

## Done criteria

- [ ] A read-only request performs no session store `Write` (or only an interval touch) — test asserts via a counting fake handler.
- [ ] GC no longer runs inline on the request goroutine; `-race` clean; the background sweeper stops on shutdown.
- [ ] `pnpm exec vp run go:test` exits 0; no out-of-scope files modified; `plans/README.md` row for 017 updated.

## STOP conditions

- Sliding-expiry semantics require a write on every request by design and the owner does not want an interval-touch approximation — report; the dirty-skip may not be acceptable as-is.
- The background GC goroutine can't be tied to a clear lifecycle (no place to stop it) — report rather than leaking a goroutine.
- Excerpts don't match live code (drift); a verification fails twice.

## Maintenance notes

- Document the interval-touch behavior and the background-GC model; a future handler must implement GC in a way that's safe to run off-request.
- Coordinate with plan 010 (also edits session `store.go`/`encrypted_store.go`) to avoid merge conflicts — land 010 first if both are scheduled.

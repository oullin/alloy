# Plan 002: Give the redis queue driver at-least-once job visibility (or explicitly document at-most-once)

> **Executor instructions**: Follow this plan step by step. Run every verification command and confirm the expected result before continuing. If a STOP condition occurs, stop and report. Update the status row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/queue/drivers/redis.go`
> If it changed, reconcile the excerpts below against live code before proceeding; on mismatch, STOP.

## Status

- **Priority**: P1
- **Effort**: L
- **Risk**: MED
- **Depends on**: `plans/001-queue-reliability.md` (attempt-tracking envelope)
- **Category**: bug
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

The redis driver's `Pop` does a plain `RPop` (which removes the job from the list) and sets `deleteFunc` to a no-op; there is no reserved set and `ReservedSize`/`ReservedJobs` return `0`/`ErrNotSupported`. If the worker dies or panics between `RPop` and completion, the job is **gone** — no redelivery. Every other production-grade queue (and this codebase's own database driver) provides visibility-timeout redelivery. This is a data-loss guarantee for a driver sold as production-ready.

This is a design decision as much as a fix: either implement a reserved sorted-set + migrate-expired step (Laravel's model), or make an explicit, documented product decision that the redis driver is at-most-once. This plan defaults to implementing reservation but includes the decision gate.

## Current state

- `pkg/hub/queue/drivers/redis.go` — `Pop` (199-243) `RPop`s and no-ops delete; `ReservedSize`/`ReservedJobs` (269-271, 369-371) return `0`/`ErrNotSupported`. The release/fail/delete closures capture the per-Pop `ctx`, so on graceful shutdown (ctx cancelled) a release/fail can itself fail.
- Contrast `pkg/hub/queue/drivers/database.go` — reserves via `reserved_at` and (after plan 001) reclaims via `retry_after`.

Excerpt (`redis.go:199-217`):

```go
func (d *RedisDriver) Pop(ctx context.Context, queueName string) (queue.Job, error) {
	d.migrateDue(ctx, queueName)
	raw, err := d.client.RPop(ctx, d.queueKey(queueName))
	if err != nil || raw == "" { return nil, queue.ErrNoJob }
	job := &redisJob{ BaseJob: BaseJob{ payload: []byte(raw), queue: queueName, connection: d.connection } }
	job.deleteFunc = func() error { return nil }
	...
}
```

Note `d.migrateDue` already exists (for delayed jobs) — a `migrateExpired` for reservations would mirror it.

Conventions: redis access is via the `d.client` abstraction (read its interface in `redis.go` / the redis client wrapper). Keys are namespaced via `d.queueKey`/`d.failedKey` helpers — add a `d.reservedKey` in the same style. Sorted sets: use the client's ZADD/ZRANGEBYSCORE equivalents; if the client interface lacks them, that is a STOP (interface extension is a bigger change to scope explicitly).

## Commands you will need

| Purpose          | Command                             | Expected |
| ---------------- | ----------------------------------- | -------- |
| Go tests (queue) | `cd pkg/hub && go test ./queue/...` | exit 0   |
| Build            | `cd pkg/hub && go build ./...`      | exit 0   |
| Full Go suite    | `pnpm exec vp run go:test`          | exit 0   |
| Format           | `pnpm exec vp run format`           | exit 0   |

## Scope

**In scope**: `pkg/hub/queue/drivers/redis.go` and its test file; a new `reservedKey` helper. If reservation is chosen, the redis client interface may need ZADD/ZRANGEBYSCORE/ZREM methods — only if they already exist or can be added without touching unrelated drivers.

**Out of scope**: other drivers; the worker loop; the payload envelope (owned by plan 001). Do not change `migrateDue`'s delayed-job behavior.

## Git workflow

- Branch: `advisor/002-redis-visibility`
- Commit per step; conventional-commit style (`fix(queue): reserve redis jobs with visibility timeout`).

## Steps

### Step 0 (decision gate): reservation vs. documented at-most-once

Confirm with the plan owner (or, if running autonomously, default to **reservation**). If the redis client interface does not expose sorted-set operations and cannot be extended within this plan's scope, fall back to the **document** path (step 3b) and STOP-report that reservation was not feasible.

### Step 1: Add a reserved sorted-set on Pop

On `Pop`, instead of `RPop` + no-op delete: atomically move the job from the list to a reserved sorted-set keyed by `d.reservedKey(queueName)` with score = now + visibility timeout (a new driver config, default e.g. 60s, mirroring the DB `retry_after`). Prefer a Lua script or `RPOPLPUSH`-style atomic move if the client supports it; otherwise document the small race window. Set `deleteFunc` to `ZREM` the job from the reserved set (a real delete, not a no-op), and `releaseFunc`/`failFunc` to remove from reserved and re-push/record accordingly.

**Verify**: test that a popped-but-not-deleted job remains recoverable; `cd pkg/hub && go test ./queue/drivers -run RedisReserve` → pass.

### Step 2: Add migrateExpired reclaim

Add `migrateExpired(ctx, queueName)` (mirroring `migrateDue`) that moves reserved entries whose score ≤ now back onto the ready list, and call it at the top of `Pop` alongside `migrateDue`. Implement `ReservedSize`/`ReservedJobs` to read the reserved sorted-set instead of returning `0`/`ErrNotSupported`.

**Verify**: test that a job left reserved past its timeout is re-popped; `go test ./queue/drivers -run RedisReclaim` → pass.

### Step 3: Decouple lifecycle closures from the per-Pop ctx

The release/fail/delete closures should not fail solely because the Pop ctx was cancelled at shutdown. Use the driver's long-lived context (or `context.Background()` with a bounded timeout) for the cleanup operations, matching how the database driver's closures behave.

**Verify**: `go test ./queue/drivers -run RedisShutdown` → pass.

### Step 3b (only if the decision was "document"):

Skip steps 1–2. Add a doc comment on `RedisDriver` and a note in `pkg/hub/queue/doc.go` stating the redis driver is at-most-once (jobs may be lost on worker crash) and pointing consumers needing at-least-once to the database or SQS drivers. Make `ReservedSize`/`ReservedJobs` documented no-ops.

### Step 4: Full suite + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

- `redis_test.go`: reservation on Pop; reclaim after timeout; delete removes from reserved set; release returns job to ready with incremented attempts (from plan 001); shutdown ctx does not break cleanup. Model after `database_integration_test.go`.

## Done criteria

- [ ] `pnpm exec vp run go:test` exits 0 with new redis reservation/reclaim tests passing (reservation path), OR the driver doc explicitly states at-most-once (document path).
- [ ] `grep -n "func() error { return nil }" pkg/hub/queue/drivers/redis.go` no longer matches the delete closure (reservation path).
- [ ] No out-of-scope files modified (`git status`).
- [ ] `plans/README.md` row for 002 updated.

## STOP conditions

- Redis client interface lacks sorted-set ops and cannot be extended in scope → take the document path and report.
- Atomic list→reserved move is not expressible with the available client primitives (report the race exposure before proceeding).
- Excerpts don't match live code (drift), or a verification fails twice.

## Maintenance notes

- The visibility-timeout default interacts with long-running handlers: too short → double processing, too long → slow reclaim. Document the tradeoff and make it configurable.
- Reviewer should confirm the list→reserved→ready transitions are atomic enough to avoid job duplication or loss under concurrent workers.

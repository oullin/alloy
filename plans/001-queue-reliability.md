# Plan 001: Queue drivers track attempts, recover from handler panics, and never silently lose failed jobs

> **Executor instructions**: Follow this plan step by step. Run every verification command and confirm the expected result before moving to the next step. If anything in the "STOP conditions" section occurs, stop and report — do not improvise. When done, update the status row for this plan in `plans/README.md`.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/queue`
> If any file under `pkg/hub/queue` changed since this plan was written, compare the "Current state" excerpts below against the live code before proceeding; on a mismatch, treat it as a STOP condition.

## Status

- **Priority**: P1
- **Effort**: L
- **Risk**: MED
- **Depends on**: none
- **Category**: bug
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

The queue worker decides when to stop retrying a job (`markIfExhausted`, `shouldFail`) purely from `job.Attempts()`. Only the **database** driver increments attempts; the redis, beanstalkd, and SQS drivers always report `Attempts() == 0`, so `max-tries`/`max-exceptions` never trip and a permanently-failing ("poison") job is released and re-processed forever, pinning a worker. Separately, a panicking job handler unwinds the worker goroutine (no `recover()`), and the database driver deletes a job from `jobs` even when writing it to `failed_jobs` fails — permanent job loss with no failed-store record. These are reliability guarantees a commercial queue primitive must hold.

This plan makes attempt tracking uniform across drivers, adds panic recovery around handler execution, and makes the database driver's fail path atomic and correctly keyed. Redis at-least-once visibility (reservation set) is a larger change handled in **plan 002**; this plan's redis change is limited to attempt tracking on the existing envelope.

## Current state

Files and roles:

- `pkg/hub/queue/worker.go` — the worker Run loop; `processJob` (line 526) calls `handler.Handle` (line 552) inline with no recover; `markIfExhausted` (579) and `shouldFail` (591) gate on `job.Attempts()`.
- `pkg/hub/queue/drivers/base_job.go` — `BaseJob` struct (line 9) with `attempts int` and the `releaseFunc`/`failFunc`/`deleteFunc` closures.
- `pkg/hub/queue/drivers/database.go` — the only driver that increments attempts (in the Pop `UPDATE ... attempts=attempts+1`); its `failFunc` is buggy.
- `pkg/hub/queue/drivers/redis.go`, `beanstalkd.go`, `sqs.go` — drivers that never populate `attempts`.
- `pkg/hub/queue/drivers/sync.go` — `SyncDriver.executeJob` runs the handler inline (line ~85), also unguarded.
- `pkg/hub/queue/drivers/failover.go` — covered by **plan 001 step 5** (error surfacing) since it is small and in the same reliability theme.

Key excerpts (as they exist today):

`worker.go:526-556` — no recover around the handler:

```go
func (w *Worker) processJob(ctx context.Context, job Job) {
	if job.IsDeleted() { ... return }
	if err := w.markIfExhausted(job); err != nil {
		w.handleJobException(job, err)
		return
	}
	...
	if err := w.handler.Handle(jobCtx, job); err != nil {
		w.handleJobException(job, err)
		return
	}
	w.emit(JobProcessed{...})
}
```

`worker.go:579-604` — attempt gating:

```go
func (w *Worker) markIfExhausted(job Job) error {
	if max := w.effectiveMaxTries(job); max > 0 && job.Attempts() > max {
		return NewMaxAttemptsExceededErrorForJob(jobNameShim{job: job})
	}
	if retryUntil := job.RetryUntil(); retryUntil != nil && time.Now().After(*retryUntil) { ... }
	return nil
}
func (w *Worker) shouldFail(job Job) bool {
	if max := w.effectiveMaxTries(job); max > 0 && job.Attempts() >= max { return true }
	...
	if job.MaxExceptions() > 0 && job.Attempts() >= job.MaxExceptions() { return true }
	return false
}
```

`drivers/redis.go:199-227` — Pop builds a job with `attempts` defaulting to 0 and re-pushes the identical raw payload on release:

```go
job := &redisJob{ BaseJob: BaseJob{ payload: []byte(raw), queue: queueName, connection: d.connection } }
job.deleteFunc = func() error { return nil }
job.releaseFunc = func(delay time.Duration) error {
	if delay > 0 { _, err := d.PushDelayed(ctx, queueName, []byte(raw), delay); return err }
	_, err := d.Push(ctx, queueName, []byte(raw)); return err
}
```

`drivers/beanstalkd.go:127-150` — `attempts` never set; beanstalkd exposes a per-job `reserves` stat. `drivers/sqs.go:197-211` — `attempts` never set; SQS exposes `ApproximateReceiveCount` on the received message.

`drivers/database.go:208-222` — the buggy fail path (uuid always empty, insert error discarded, delete unconditional):

```go
job.failFunc = func(err error) error {
	var errMsg string
	if err != nil { errMsg = err.Error() }
	errBytes, _ := json.Marshal(map[string]string{"exception": errMsg})
	_ = d.db.Exec(ctx,
		"INSERT INTO failed_jobs (uuid, connection, queue, payload, exception) VALUES ($1,$2,$3,$4,$5)",
		job.uuid, d.connection, queueName, payload, string(errBytes),
	)
	return job.deleteFunc()
}
```

The `dbJob` built at `database.go:187-194` sets only `id/payload/queue/attempts` — `uuid` is empty. The `failed_jobs.uuid` column is `TEXT NOT NULL UNIQUE` (schema comment at `database.go:29-37`).

`drivers/failover.go:116-126` — Pop swallows all backend errors:

```go
func (d *FailoverDriver) Pop(ctx context.Context, queueName string) (queue.Job, error) {
	for _, drv := range d.drivers {
		job, err := drv.Pop(ctx, queueName)
		if err == nil && job != nil { return job, nil }
	}
	return nil, queue.ErrNoJob
}
```

Repo conventions to match:

- Errors: wrap with `fmt.Errorf("...: %w", err)` when carrying an underlying error (this is the convention the typed-errors work converged on). Sentinel errors live in `pkg/hub/queue/errors.go` (e.g. `queue.ErrNoJob`).
- The `attempts` field is a plain `int`; the payload envelope is JSON (`pkg/hub/queue/payload.go`). Check whether the JSON envelope already carries a `tries`/`attempts` field before inventing one — read `payload.go` and `payload_builder.go`.
- Tests are table-driven; driver tests use fakes. See `pkg/hub/queue/drivers/database_integration_test.go` and `pkg/hub/queue/worker_test.go` for the established patterns. UUIDs use `github.com/google/uuid` or `github.com/oklog/ulid/v2` (both already in go.mod — check which the codebase uses for job identity via `grep -rn "uuid.New\|ulid.Make" pkg/hub/queue`).

## Commands you will need

| Purpose                            | Command                             | Expected on success |
| ---------------------------------- | ----------------------------------- | ------------------- |
| Go tests (all modules)             | `pnpm exec vp run go:test`          | exit 0, all pass    |
| Go tests (queue only, faster loop) | `cd pkg/hub && go test ./queue/...` | exit 0              |
| Typecheck/build                    | `cd pkg/hub && go build ./...`      | exit 0              |
| Format                             | `pnpm exec vp run format`           | exit 0              |

## Scope

**In scope** (modify only these):

- `pkg/hub/queue/worker.go` (panic recovery)
- `pkg/hub/queue/drivers/sync.go` (panic recovery in executeJob)
- `pkg/hub/queue/drivers/redis.go`, `beanstalkd.go`, `sqs.go` (attempt tracking)
- `pkg/hub/queue/drivers/database.go` (fail-path integrity)
- `pkg/hub/queue/drivers/failover.go` (error surfacing)
- `pkg/hub/queue/drivers/base_job.go` (only if a setter/field is genuinely needed)
- Corresponding `_test.go` files in the same directories (create where missing)

**Out of scope** (do NOT touch):

- Redis reservation/visibility-timeout redesign — that is **plan 002**. Here, redis gets attempt tracking only, on the existing RPop envelope.
- The `queue.Job` interface shape / `payload.go` envelope format beyond adding an attempt count if one isn't already present.
- `worker.go` sleep/backoff behavior (that is plan 017/P11 territory).

## Git workflow

- Branch: `advisor/001-queue-reliability`
- Commit per step; message style matches `git log` (e.g. `fix(queue): track attempts in redis driver`).
- Do NOT push or open a PR unless the operator instructs it.

## Steps

### Step 1: Add panic recovery around handler execution

In `worker.go`, wrap the `w.handler.Handle(jobCtx, job)` call (line ~552) so a panic is converted into an error routed through `w.handleJobException(job, err)` instead of unwinding the goroutine. Use a deferred recover in a small helper, e.g.:

```go
func (w *Worker) runHandler(ctx context.Context, job Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("queue: handler panicked: %v", r)
		}
	}()
	return w.handler.Handle(ctx, job)
}
```

Call `w.runHandler(jobCtx, job)` in `processJob`. Apply the same recover wrapper around the handler/`Fire` call in `SyncDriver.executeJob` (`drivers/sync.go`, line ~85, `job.Fire(ctx)`).

**Verify**: add a test in `worker_test.go` with a handler that panics; assert the worker does not crash and the job is failed/released via `handleJobException`. `cd pkg/hub && go test ./queue/ -run Panic` → pass.

### Step 2: Determine the attempt-count source of truth

Read `pkg/hub/queue/payload.go` and `payload_builder.go`. Decide: does the JSON envelope already carry an attempt/`tries` count?

- **If yes**: parse it into `BaseJob.attempts` on each driver's Pop and increment-and-persist it on release.
- **If no**: add a `tries` field to the envelope (default 0), and have every driver read it on Pop and write the incremented value back on release.

Record the decision in a one-line comment in `payload.go`. **If the envelope format is shared with an external system (check for a schema/version marker), STOP and report** — changing the wire format may break compatibility.

**Verify**: `cd pkg/hub && go build ./queue/...` → exit 0.

### Step 3: Track attempts in redis, beanstalkd, and SQS Pop

- **redis** (`redis.go:199-227`): parse the attempt count from the payload envelope (per step 2) into `BaseJob.attempts`; on `releaseFunc`, re-push the payload with the incremented count rather than the identical `raw`.
- **beanstalkd** (`beanstalkd.go:127`): set `attempts` from the tube job's `reserves` stat (fetch via the client's stats-job call; if the client interface lacks it, read the envelope count as the fallback used by redis).
- **sqs** (`sqs.go:197`): set `attempts` from `ApproximateReceiveCount` on the received message (it is an SQS system attribute — ensure the receive request asks for it).

**Verify**: driver-level tests asserting that a job popped after N releases reports `Attempts() == N`. `cd pkg/hub && go test ./queue/drivers/ -run Attempts` → pass.

### Step 4: Fix the database driver fail path

In `database.go`'s `failFunc` (208-222):

1. Populate a real `uuid` for the job at Pop time (generate with the same library the codebase uses for job identity — see conventions). Set it on the `dbJob.BaseJob` at `database.go:187-194`.
2. Propagate the `INSERT INTO failed_jobs` error instead of discarding it (`_ =`).
3. Only call `job.deleteFunc()` after the failed-store insert succeeds. Prefer wrapping both statements in a transaction if the `DBExecer` interface supports one; otherwise order them insert-then-delete and return early on insert error.

**Verify**: a test where the failed_jobs insert errors (fake `DBExecer` returning an error on the insert) asserts the job is NOT deleted and the error is returned. `cd pkg/hub && go test ./queue/drivers/ -run Fail` → pass.

### Step 5: Surface real errors in the failover driver

In `failover.go`, `Pop` (116-126): distinguish `queue.ErrNoJob` from real backend errors. Track the last real error; if every backend returned a real (non-`ErrNoJob`) error, return that wrapped error rather than `ErrNoJob`. Emit the existing `FailedOver` event when falling through from one backend to the next (check `events.go` for the event type). Change the `Size`/`PendingSize`/`DelayedSize`/`ReservedSize` family (128-174) to return the last real error instead of `0, nil` when all backends fail.

**Verify**: a test with two fake drivers both returning errors asserts `Pop` returns a non-nil error (not `ErrNoJob`) and `Size` returns the error. `cd pkg/hub && go test ./queue/drivers/ -run Failover` → pass.

### Step 6: Full queue suite + format

**Verify**: `pnpm exec vp run go:test` → exit 0, all pass. `pnpm exec vp run format` → exit 0.

## Test plan

New/updated tests (table-driven, matching `worker_test.go` and `database_integration_test.go` style):

- `worker_test.go`: panicking handler does not crash the worker; job routed to exception handling. Handler returning error still respects max-tries once attempts are tracked.
- `drivers/redis_test.go` (create if absent), `beanstalkd_test.go`, `sqs_test.go`: `Attempts()` increments across release/redelivery.
- `drivers/database_test.go` (or extend the integration test): fail path does not delete when failed_jobs insert errors; uuid is non-empty; a poison job is moved to failed_jobs after max-tries instead of looping.
- `drivers/failover_test.go`: all-backends-error surfaces an error, not empty.

## Done criteria

ALL must hold:

- [ ] `pnpm exec vp run go:test` exits 0; new tests above exist and pass.
- [ ] `grep -n "_ = d.db.Exec" pkg/hub/queue/drivers/database.go` returns nothing in the failFunc (error is handled).
- [ ] `grep -n "attempts:" pkg/hub/queue/drivers/redis.go pkg/hub/queue/drivers/beanstalkd.go pkg/hub/queue/drivers/sqs.go` shows attempts sourced on Pop (or the envelope-parse equivalent).
- [ ] A panicking handler test passes without crashing the test binary.
- [ ] No files outside the in-scope list are modified (`git status`).
- [ ] `plans/README.md` status row for 001 updated.

## STOP conditions

Stop and report (do not improvise) if:

- The payload envelope has an external schema/version marker (step 2) — wire-format changes need owner sign-off.
- The beanstalkd or SQS client interface exposes no way to read the redelivery count and no envelope count exists to fall back on.
- The `DBExecer` interface cannot express a transaction and you cannot make insert-then-delete safe without one.
- Any "Current state" excerpt does not match the live code (drift).
- A verification fails twice after a reasonable fix attempt.

## Maintenance notes

- Plan 002 (redis visibility) builds directly on the attempt envelope from step 2 — keep the envelope field name stable.
- Reviewer should scrutinize: the redelivery-count semantics per backend (SQS `ApproximateReceiveCount` counts _receives_, which includes visibility-timeout expiries — confirm that matches the intended "attempts" meaning), and that the database fail path is genuinely atomic.
- Deferred out of this plan: adaptive worker poll backoff (P11) and redis reservation/visibility (plan 002).

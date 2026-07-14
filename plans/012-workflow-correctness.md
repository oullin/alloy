# Plan 012: Workflow apply concurrency/ordering (Go) and enforce or remove `maxExceptions` (TS)

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/workflow sdk/workflow`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: bug
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

1. **C17 (Go)**: `Machine.Apply` reads the marking, validates the guard, then writes the new marking with no lock, version, or CAS spanning the read and write. Two goroutines applying transitions to the same subject can both read the same marking, both pass the guard, and both write — a lost update that double-consumes tokens or produces an impossible Petri-net marking. Separately, `Apply` dispatches the leave/transition/enter events **before** `SetMarking`, so those events fire even when the write subsequently fails (phantom transitions).
2. **C20 (TS)**: `RetryPolicy.maxExceptions` is declared, stored, and threaded from job config, but `runWithRetry` never reads it — the retry loop only bounds on `maxTries`/`signal.aborted`. A documented retry knob is a silent no-op (and likely diverges from the Go twin).

## Current state

- `pkg/hub/workflow/apply.go:6-38` — `Apply`:
  ```go
  transition, next, err := w.prepareApply(subject, transitionName, context) // reads GetMarking, checks guard
  if err != nil { ... return Marking{}, err }
  w.dispatchLeaveEvents(subject, transition, next, context)       // fired BEFORE the write
  w.dispatchTransitionEvents(subject, transition, next, context)
  w.dispatchEnterEvents(subject, transition, next, context)
  if err := w.store.SetMarking(subject, next, w.definition, context); err != nil { ... return Marking{}, err }
  w.dispatchEnteredEvents(...)
  ...
  ```
  `prepareApply` (below) calls `w.GetMarking(subject)`. There is no CAS/version on `SetMarking`.
- `sdk/workflow/src/multisteps/retry.ts:5-11` declares/stores `maxExceptions`; `jobs.ts:52` threads it in; `runWithRetry` (23-58) never reads it.

Convention: the Go workflow is the twin of `sdk/workflow`; keep semantics aligned. The `MarkingStore` is an interface — adding optimistic concurrency means either a versioned `SetMarking` (reject if the marking changed since read) or a documented "callers serialize per subject" contract.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Go workflow tests | `cd pkg/hub && go test ./workflow/...` | exit 0 |
| Go race | `cd pkg/hub && go test -race ./workflow/...` | exit 0 |
| TS workflow tests | `pnpm exec vp test` (workflow) | pass |
| Full Go suite | `pnpm exec vp run go:test` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `pkg/hub/workflow/apply.go` (event ordering; concurrency), the `MarkingStore` contract if a CAS/version is added, `sdk/workflow/src/multisteps/retry.ts` (+ `jobs.ts`/`engine.ts` as needed for `maxExceptions`), and the corresponding tests.

**Out of scope**: workflow definition/DSL; other multistep engine behavior (the `continueOnError` question C23/TS-14 is a separate investigate item — note it but don't fix here unless trivial).

## Git workflow

- Branch: `advisor/012-workflow-correctness`; commit per concern; conventional-commit style.

## Steps

### Step 1 (C17a): Move event dispatch after a successful write

Reorder `Apply` so the leave/transition/enter events are dispatched only after `SetMarking` succeeds (or adopt a clear two-phase model where pre-write events are explicitly "about to transition" and post-write events are "transitioned"). At minimum, no event that asserts a completed transition may fire before the write succeeds. Preserve the existing event *types* and their payloads.

**Verify**: `go test ./workflow -run Events` → a test with a failing `SetMarking` asserts the enter/entered events did **not** fire.

### Step 2 (C17b): Add optimistic concurrency to Apply

Choose one, matching the twin's intended contract:
- **(a) Versioned CAS**: add a version/token to the marking; `SetMarking` rejects the write if the stored marking changed since `GetMarking`. `Apply` retries or returns a conflict error.
- **(b) Documented serialization**: if per-subject serialization is the intended model, document that `MarkingStore` implementations/callers must serialize per subject, and add a note to `Apply`.

Prefer (a) if the `MarkingStore` interface can carry a version without breaking existing implementers; otherwise (b). **If (a) requires changing an externally-implemented `MarkingStore`, STOP and report** the blast radius first.

**Verify**: `go test -race ./workflow -run Concurrent` → two concurrent `Apply` calls on the same subject don't both succeed into an impossible marking (one wins, the other retries or errors).

### Step 3 (C20): Enforce or remove `maxExceptions` (TS)

Decide against the Go twin's semantics. Either implement the exception-count cap in `runWithRetry` (stop retrying once accumulated exceptions reach `maxExceptions`) or remove the option entirely. If implementing, mirror how the Go `shouldFail`/`MaxExceptions` gate behaves.

**Verify**: `pnpm exec vp test` (workflow) → a policy with `maxExceptions: N` stops after N exceptions (or the option is gone and no test/usage references it).

### Step 4: Full suites + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp test` → pass; `pnpm exec vp run format` → exit 0.

## Test plan

- Go `workflow` tests: event ordering (no enter/entered before a successful write); concurrent `Apply` on one subject → no lost update (under `-race`).
- TS `workflow` tests: `maxExceptions` enforced (or removed and unreferenced).

## Done criteria

- [ ] `go test -race ./workflow/...` exits 0; concurrent-apply test present.
- [ ] Enter/entered events do not fire when `SetMarking` fails (test asserts it).
- [ ] `sdk/workflow` `maxExceptions` is either enforced (test asserts) or removed (`grep -rn maxExceptions sdk/workflow/src` returns nothing).
- [ ] `pnpm exec vp run go:test` and `pnpm exec vp test` pass; no out-of-scope files modified; `plans/README.md` row for 012 updated.

## STOP conditions

- Adding a version to `MarkingStore` breaks external implementers — take the documented-serialization path and report.
- The Go and TS twins turn out to intend *different* `maxExceptions` semantics — report the divergence rather than guessing.
- Excerpts don't match live code (drift); a verification fails twice.

## Maintenance notes

- Note the `continueOnError`/`runAsyncLenient` question (TS-14): the lenient path still throws on the first async error — flagged as a separate investigate item, not fixed here.
- Reviewer should confirm event ordering didn't drop any event that a consumer relies on firing pre-write.

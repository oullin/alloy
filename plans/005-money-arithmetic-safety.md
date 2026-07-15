# Plan 005: Overflow-safe money Add/Subtract, fixed Absolute guard, and one consistent rounding policy (Go + TS)

> **REVISION 2026-07-15 (reviewer reconciliation — this section OVERRIDES the stale parts of the plan below).**
> Executor verification against live code (and against `bfface5` itself) found the original "Current state" partially wrong. Corrections, authoritative:
>
> 1. **Naming collision (Step 1)**: `pkg/hub/money/calculator/ration.go:10,19` ALREADY defines package funcs `SafeAdd(a, b Amount) Amount` / `SafeSubtract(a, b Amount) Amount` that silently return 0 on overflow; they back `Engine.Add`/`Engine.Subtract`, which `Manager.Split`/`Allocate` use. DO NOT redefine or change them. Instead add **methods** `Engine.SafeAdd(a, b Amount) (Amount, error)` and `Engine.SafeSubtract(a, b Amount) (Amount, error)` mirroring `Engine.SafeMultiply` (calculator.go:32), returning `exception.ErrOverflow`. `Manager.Add`/`Subtract` (manager.go:127,150) route through these new methods exactly as `Multiply` does. Leave `Split`/`Allocate` and the ration.go funcs untouched — their silent-0 behavior is recorded as a separate follow-up finding, out of this plan's scope.
> 2. **Step 2 TS half is already done**: TS `add`/`subtract` already route through throwing `safeAdd`/`safeSubtract`. TS work in step 2 is verification only (a passing test may be added, no production change).
> 3. **Step 3 (Absolute)**: Go fix stands — replace the always-false guard so `Absolute(math.MinInt64)` returns `0` (matching the package's existing silent-0 overflow convention in ration.go; document it on the method). TS `absolute` has NO MinInt64 guard and bigint can represent 2^63; for cross-runtime parity add a guard returning `0n` when the input is `MIN_INT64` (i.e. when negation exits int64 range), documented the same way. Plan 008's fixtures will lock this pair.
> 4. **Step 4**: Go `calculator.Round` and TS `round` tie-handling changes stand (half away from zero — pre-approved). TS `aggregator.avg` truncation fix stands. There is NO Go `Aggregator.Avg` — ignore the plan's instruction to align it; do not add one.
> 5. Everything else (verification commands, scope, STOP conditions, git workflow) stands. Scope explicitly includes `pkg/hub/money/calculator/ration.go` ONLY if a doc comment is added there; its behavior must not change.

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/money sdk/money`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: bug
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

This library exists to make money math safe, yet three defects silently produce wrong amounts:
1. `Manager.Add`/`Subtract` use raw `+=`/`-=`, so summing large values wraps int64 with no error — while `Multiply` correctly routes through `SafeMultiply` and returns `ErrOverflow`. `Aggregator.Sum`/`Avg` build on the unsafe path.
2. `calculator.Absolute`'s overflow guard `amount < math.MinInt64` is always false, so `Absolute(math.MinInt64)` returns a negative number, poisoning `Round` and `Split`.
3. Rounding is inconsistent: `calculator.Round` (Go and TS) rounds exact halves toward zero, but the TS doc says "half away from zero" and `CreateFromFloat` uses `math.Round` (half away). The same library rounds exact halves two different ways depending on entry point.

## Current state

- `pkg/hub/money/money/manager.go`:
  - `Add` (110-131): `result.amount += m2.amount` at line 127.
  - `Subtract` (133-154): `result.amount -= m2.amount` at line 150.
  - `Multiply` (156-173): correctly uses `mm.calculator.SafeMultiply` (166) and returns its error.
  - `CreateFromFloat` (57-64): `mm.Create(int64(math.Round(scaled)), code)`.
- `pkg/hub/money/calculator/calculator.go`:
  - `Add(a, b Amount) Amount` exists (line 17) but there is **no `SafeAdd`/`SafeSubtract`** returning an error (only `SafeMultiply`, line 32).
  - `Absolute` (67-78): guard `if c == nil || amount < math.MinInt64` (line 69, always false); returns `-amount` for `math.MinInt64` (still negative).
  - `Round` (93-131): tie rule `if module > (reminder / 2)` (line 120) → ties toward zero.
- `pkg/hub/money/exception/exception.go`: `ErrOverflow` (42), `ErrNoMultipliersProvided` (27) — reuse `ErrOverflow`.
- TS twin `sdk/money/src/calculator.ts`:
  - `safeAdd`/`safeSubtract` (115-133) already exist and `throw ERR_OVERFLOW` when `!inInt64Range(result)`.
  - `round` (95-113): doc says "half away from zero" (line 95) but code is `if (module > reminder / 2n)` (106) → ties toward zero (**doc/impl mismatch**).
  - `absolute`: verify the same `MinInt64` guard bug exists and fix to match Go.
- TS `sdk/money/src/money/manager.ts` / `aggregator.ts`: confirm `add`/`subtract` route through `safeAdd`/`safeSubtract`; `MoneyAggregator.avg` truncates (handled in this plan's step 4, shared with finding C19).

Excerpt (`calculator.go:67-78`):
```go
func (c *Engine) Absolute(amount Amount) Amount {
	if c == nil || amount < math.MinInt64 { return 0 }
	if amount < 0 { return -amount }
	return amount
}
```

Convention: safe arithmetic returns `(int64, error)` with `exception.ErrOverflow`; the manager propagates it exactly as `Multiply` does. Rounding-mode decision is a **product decision** — this plan defaults to **half away from zero** (matches `CreateFromFloat` and the documented TS intent) unless the owner specifies banker's rounding.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Go money tests | `cd pkg/hub && go test ./money/...` | exit 0 |
| TS money tests | `pnpm --filter @alloy/sdk/money test` (confirm exact filter via `pnpm -r ls`) or `pnpm exec vp test` | pass |
| Full Go suite | `pnpm exec vp run go:test` | exit 0 |
| Typecheck TS | `pnpm exec vp check` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**:
- `pkg/hub/money/calculator/calculator.go` (add `SafeAdd`/`SafeSubtract`, fix `Absolute`, set the tie rule)
- `pkg/hub/money/money/manager.go` (route `Add`/`Subtract` through safe ops)
- `sdk/money/src/calculator.ts` (fix `round` doc/impl to the chosen policy, fix `absolute` guard)
- `sdk/money/src/money/manager.ts` and `aggregator.ts` (confirm/route safe ops; fix `avg`)
- Corresponding test files on both sides.

**Out of scope**: currency conversion (that is **plan 006**); the `Value`/`Amount` types; `Split`/`Allocate` algorithm beyond consuming the fixed `Absolute`.

## Git workflow

- Branch: `advisor/005-money-arithmetic-safety`
- Commit per concern; conventional-commit style.

## Steps

### Step 1: Add Go `SafeAdd`/`SafeSubtract`

In `calculator.go`, add `SafeAdd(a, b int64) (int64, error)` and `SafeSubtract(a, b int64) (int64, error)` that detect int64 overflow (standard checked-add: overflow iff `(b > 0 && a > maxInt64-b) || (b < 0 && a < minInt64-b)`) and return `exception.ErrOverflow`. Mirror the existing `SafeMultiply` style. Do **not** change the existing `Add` (line 17) signature if other callers depend on it — add the new safe variants alongside.

**Verify**: `cd pkg/hub && go test ./money/calculator -run Safe` → pass; add cases for `MaxInt64 + 1` and `MinInt64 - 1` returning `ErrOverflow`.

### Step 2: Route `Manager.Add`/`Subtract` through the safe ops

In `manager.go`, replace `result.amount += m2.amount` (127) and `result.amount -= m2.amount` (150) with `SafeAdd`/`SafeSubtract`, propagating `ErrOverflow` the way `Multiply` does. `Add`/`Subtract` already return `error`, so no signature change.

**Verify**: `go test ./money/money -run 'Add|Subtract|Overflow'` → summing to overflow returns `ErrOverflow`, not a wrapped value. `Aggregator.Sum` on an overflowing set returns the error.

### Step 3: Fix `calculator.Absolute` (Go and TS)

Change the Go guard to `if c == nil || amount == math.MinInt64 { return 0 }` (or return an `(Amount, error)` if you prefer surfacing it — but keep the call sites in `Round`/`Split` working; the simplest safe fix is return 0 for the single unrepresentable input). Apply the identical fix to `sdk/money/src/calculator.ts` `absolute`.

**Verify**: `go test ./money/calculator -run Absolute` asserts `Absolute(math.MinInt64) >= 0`. TS equivalent test passes.

### Step 4: Unify the rounding policy (Go + TS) and fix `avg`

Decide the canonical mode (default: **half away from zero**). Make `calculator.Round` (Go, line 120) and `sdk/money/src/calculator.ts` `round` (106) implement it identically, and update the TS doc comment (line 95) to match the actual behavior. Ensure `CreateFromFloat` (already half-away) agrees. Fix `MoneyAggregator.avg` (`sdk/money/src/money/aggregator.ts:70`) to round the quotient with the same policy (or distribute the remainder via `split`) rather than truncating BigInt division; confirm the Go `Aggregator.Avg` uses the same rule.

**Verify**: `go test ./money/... -run Round` and the TS money round/avg tests pass, including an exact-half case (e.g. rounding `2.5`-equivalent minor units) producing the same result on both runtimes.

### Step 5: Full suites + format

**Verify**: `pnpm exec vp run go:test` → exit 0. `pnpm exec vp check` → exit 0. `pnpm exec vp test` → pass. `pnpm exec vp run format` → exit 0.

## Test plan

- Go `calculator_test.go`: `SafeAdd`/`SafeSubtract` overflow cases; `Absolute(MinInt64)`; `Round` exact-half under the chosen policy.
- Go `manager_test.go`/`aggregator_test.go`: `Add`/`Subtract`/`Sum` overflow returns `ErrOverflow`; `Avg` rounding.
- TS `calculator`/`aggregator` tests: matching `round` policy + doc, fixed `absolute`, `avg` rounding.
- Where a value is chosen to demonstrate the fix, pick the same numeric input on both runtimes so plan 008's fixtures can reuse it.

## Done criteria

- [ ] `grep -n "result.amount +=\|result.amount -=" pkg/hub/money/money/manager.go` returns nothing.
- [ ] `cd pkg/hub && go test ./money/...` exits 0; overflow and Absolute tests present.
- [ ] TS money tests pass; `sdk/money/src/calculator.ts` round doc matches its code.
- [ ] The same exact-half input rounds identically in Go and TS (a test on each side asserts the same result).
- [ ] No out-of-scope files modified; `plans/README.md` row for 005 updated.

## STOP conditions

- The owner wants banker's rounding (or another mode) rather than half-away — confirm before changing `Round`, since it shifts existing results at exact halves.
- Changing `Absolute`'s signature to return an error would ripple into many call sites — prefer the return-0 fix and report if that is unacceptable.
- Excerpts don't match live code (drift); a verification fails twice.

## Maintenance notes

- Plan 008 encodes these exact behaviors as shared Go↔TS fixtures — keep the chosen rounding policy stable or update 008.
- Reviewer should confirm `Split`/`Allocate` (which consume `Absolute`) produce the same distributions after the guard fix.

# Plan 006: Replace float-based currency conversion with bounds-checked integer math (Go + TS)

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/money/exchange sdk/money/src/exchange`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: MED
- **Depends on**: none (independent of 005, but shares the money test conventions)
- **Category**: bug
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

Currency conversion routes exact int64 minor-unit amounts through `float64` and then narrows back with an unchecked `int64(math.Round(...))`. Above 2^53 minor units the `float64(amount)` cast loses precision, and converting an out-of-range float to int64 in Go is implementation-defined (no predictable wrap) with no overflow check. The TS twin does the same via `Number(amount)`. For a library whose premise is exact money math, large-value conversions silently lose cents or produce garbage — in the one path most likely to feed real financial totals.

## Current state

- `pkg/hub/money/exchange/exchange.go`:
  - `ConvertAmount` (~105-120): `majorUnits := float64(amount) / fromFractionPow; converted := int64(math.Round(majorUnits * rate * toFractionPow))`.
  - `ConvertAmountWithRate` (123-140): same pattern; guards `rate <= 0` → `ErrInvalidExchangeRate`.
- `sdk/money/src/exchange/rates.ts`:
  - `convertAmountWithRate` (68-74): `const majorUnits = Number(amount) / 10 ** fromFraction; ... return roundAwayFromZero(convertedMajorUnits * 10 ** toFraction)`.
  - Inverse rates derived as `1 / inverse` (float reciprocal) around line 52.

Excerpt (`exchange.go` conversion core):
```go
majorUnits := float64(amount) / fromFractionPow
convertedMajorUnits := majorUnits * rate
convertedAmount := int64(math.Round(convertedMajorUnits * toFractionPow))
return convertedAmount, nil
```

Convention: money is int64 minor units; the calculator (`pkg/hub/money/calculator`) already provides `SafeMultiply` and (after plan 005, if landed) `SafeAdd`. Rates are the one inherently fractional input — represent the rate as a scaled integer (numerator/denominator or a fixed-scale integer) so the multiply/divide stays in integer space, and bounds-check before narrowing. Return `exception.ErrOverflow` on out-of-range results, matching the rest of the package.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Go exchange tests | `cd pkg/hub && go test ./money/exchange/...` | exit 0 |
| TS money tests | `pnpm exec vp test` (money package) | pass |
| Full Go suite | `pnpm exec vp run go:test` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `pkg/hub/money/exchange/exchange.go`, `sdk/money/src/exchange/rates.ts`, their tests. The rate representation (scaled-integer helper) may be added within these files.

**Out of scope**: `Manager.Add`/`Subtract`/rounding (plan 005); the currency dataset; the public `Convert` entry points' signatures (keep stable — only the internals change).

## Git workflow

- Branch: `advisor/006-money-exchange-integer`; commit per runtime; conventional-commit style.

## Steps

### Step 1: Choose the rate representation

Represent the exchange rate as a scaled integer (e.g. rate × 10^scale as an int64/bigint, with a fixed `scale` such as 12) plus explicit rounding at the end. Document the supported rate precision and the maximum representable amount. **If the public API accepts `rate float64`** (it does today), keep that signature but convert the float to the scaled integer once, at the boundary, with a documented precision — do not thread floats through the multiply.

### Step 2: Reimplement Go conversion with integer math + overflow checks

Rewrite `ConvertAmount` and `ConvertAmountWithRate` to compute `amount * rateScaled` and divide by the scale and fraction powers using integer operations (use `SafeMultiply` and bounds checks; for the division-with-rounding step apply the same rounding policy the package uses — coordinate with plan 005's decision if landed, else `math`-free half-away integer rounding). Return `exception.ErrOverflow` when the result exceeds int64 range instead of narrowing blindly.

**Verify**: `cd pkg/hub && go test ./money/exchange -run 'Convert|Overflow|Precision'` → passes, including a large-amount case that previously lost precision and an overflow case returning `ErrOverflow`.

### Step 3: Mirror in TS

Reimplement `convertAmountWithRate` (and the inverse-rate derivation) in `sdk/money/src/exchange/rates.ts` using bigint scaled-integer math; range-check against int64 bounds (reuse `inInt64Range`) and throw `ERR_OVERFLOW`. Ensure the same inputs produce the same outputs as the Go side.

**Verify**: TS exchange tests pass, including the same large-amount and overflow cases used in step 2.

### Step 4: Full suites + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp test` → pass; `pnpm exec vp run format` → exit 0.

## Test plan

- Go `exchange_test.go` and TS exchange tests: representative conversions (whole and fractional rates), a large amount (> 2^53 minor units) preserving precision, an overflow input returning/throwing the overflow error, and inverse-rate symmetry within one minor unit.
- Choose identical numeric fixtures on both runtimes so plan 008 can adopt them.

## Done criteria

- [ ] `grep -n "float64(amount)" pkg/hub/money/exchange/exchange.go` returns nothing (no float cast of the amount).
- [ ] `grep -n "Number(amount)" sdk/money/src/exchange/rates.ts` returns nothing.
- [ ] Go and TS exchange tests pass with matching large-amount results.
- [ ] Overflow returns `ErrOverflow` (Go) / throws `ERR_OVERFLOW` (TS).
- [ ] No out-of-scope files modified; `plans/README.md` row for 006 updated.

## STOP conditions

- The public conversion API must change its signature to express the rate precisely and the owner hasn't approved it — report; the float-boundary approach in step 1 should avoid this.
- Rounding at the final divide can't match plan 005's policy without ambiguity — report so the two plans align.
- Excerpts don't match live code (drift); a verification fails twice.

## Maintenance notes

- Document the rate scale and max amount; consumers with rates needing more precision than the chosen scale must be told.
- Reviewer should verify inverse-rate conversions are symmetric to within one minor unit and that rounding matches plan 005.

# Plan 007: Make TS tempo day/week arithmetic DST-correct and locale-aware month parsing

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- sdk/tempo pkg/hub/tempo`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: bug
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

Two TS tempo bugs diverge from the Go engine (the reference implementation) and produce wrong times:

1. `add()` treats `day` and `week` as fixed-millisecond units (24h/168h), so adding a day across a DST transition yields the wrong wall-clock time (a real calendar day is 23h or 25h). `startOfDay().addDays(1)` no longer lands on midnight, breaking `today()`/`tomorrow()`/`isTomorrow()` for zoned usage. Go's `internal/kernel/arithmetic.go` correctly uses calendar-aware `AddDate` for day/week; `addMonths` on the TS side is already zone-aware — day/week are the outliers.
2. `parseFromPattern` looks up month names with no locale (defaults to `en-US`), so parsing a localized formatted date (e.g. French `"14 juillet 2026"`) fails the name match and silently falls back to **January** via `Number(values.get('MM') ?? ... ?? 1)`. Round-tripping `format(locale)` → `parseFromPattern` is broken for any non-English locale.

## Current state

- `sdk/tempo/src/core/index.ts`:
    - `add()` (1058-1066): `const fixed = fixedUnitMilliseconds(unit); if (fixed !== null) { return this.make(new Date(this.value.getTime() + value * fixed)); }` then a switch handling `month` via the zone-aware `addMonths`.
    - `addDays` (~1144), `addWeeks` (~1179) delegate to `add`.
- `sdk/tempo/src/calendar/index.ts`:
    - `fixedUnitMilliseconds` (300-322): returns constants for `millisecond/second/minute/hour/day/week` — `day` and `week` should not be here.
    - `monthNumberFromName` (~365): defaults locale to `en-US`.
- `sdk/tempo/src/parsing/index.ts:296`: `month: monthNumberFromName(values.get('MMMM') ?? values.get('MMM') ?? '') ?? Number(values.get('MM') ?? values.get('M') ?? 1)` — no locale passed, silent January fallback.
- Go reference: `pkg/hub/tempo/internal/kernel/arithmetic.go` handles `duration.Day`/`Week` via calendar `AddDate` (DST-aware). Confirm `addMonths` in `sdk/tempo/src/core/index.ts` (~1187-1205) uses `dateFromZonedComponents` — that is the pattern to reuse for day/week.

Convention: zoned arithmetic goes through the existing zoned-component helpers (`dateFromZonedComponents` / the calendar module), not raw `Date.getTime()` math. Options carry `locale`/`timeZone` and are already threaded elsewhere in tempo.

## Commands you will need

| Purpose        | Command                                                                          | Expected |
| -------------- | -------------------------------------------------------------------------------- | -------- |
| TS tempo tests | `pnpm exec vp test` (tempo) or `pnpm --filter @alloy/sdk/tempo-tests test:tempo` | pass     |
| Typecheck      | `pnpm exec vp check`                                                             | exit 0   |
| Format         | `pnpm exec vp run format`                                                        | exit 0   |

## Scope

**In scope**: `sdk/tempo/src/core/index.ts`, `sdk/tempo/src/calendar/index.ts`, `sdk/tempo/src/parsing/index.ts`, and the tempo test suite (`sdk/tempo/tests/`).

**Out of scope**: the Go tempo package (it is the correct reference — do not change it); tempo's structural refactor (deferred X10); formatting paths (perf plan 016).

## Git workflow

- Branch: `advisor/007-tempo-ts-dst-and-locale`; commit per concern; conventional-commit style.

## Steps

### Step 1: Remove day/week from fixed-millisecond units

In `calendar/index.ts`, make `fixedUnitMilliseconds` return non-null only for `millisecond/second/minute/hour` — return `null` for `day`, `week` (and anything larger). This forces `add()` into the calendar-aware branch.

**Verify**: `pnpm exec vp check` → exit 0 (no type errors from the switch exhaustiveness).

### Step 2: Add calendar-aware day/week to `add()`

In `core/index.ts` `add()`, handle `day` and `week` in the switch using the same zoned-component approach as `addMonths` (shift the day component in the target time zone via `dateFromZonedComponents`), so a day add lands on the same wall-clock time the next calendar day, DST included. `addWeeks` = 7×day.

**Verify**: add DST regression tests (see Test plan). `pnpm exec vp test` (tempo) → pass, including a zone like `America/New_York` across a spring-forward boundary where `startOfDay().addDays(1)` equals the next local midnight.

### Step 3: Thread locale into month-name parsing

In `parsing/index.ts:296`, pass the parse `options?.locale` (or the policy locale) into `monthNumberFromName`. When a month-name token is present but matches nothing, do not silently fall back to January — in strict mode throw/return an error; otherwise document the fallback explicitly. Ensure `monthNumberFromName` accepts and uses the locale rather than defaulting to `en-US` when one is supplied.

**Verify**: a test parsing a French `"14 juillet 2026"` with locale `fr-FR` and pattern `D MMMM YYYY` yields month 7, and an unmatched name does not silently become January. `pnpm exec vp test` (tempo) → pass.

### Step 4: Full suite + format

**Verify**: `pnpm exec vp test` → pass; `pnpm exec vp check` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

- `sdk/tempo/tests/…`:
    - DST: in a DST-observing zone, `addDays(1)` across spring-forward and fall-back lands on the correct local wall-clock time; `startOfDay().addDays(1)` is next local midnight; `today()/tomorrow()/isTomorrow()` correct in that zone.
    - UTC unchanged: `addDays`/`addWeeks` in UTC produce identical results to before (no regression).
    - Locale parse: non-English month name round-trips `format(locale)` → `parseFromPattern`; unmatched name is not silently January.
- Where practical, choose inputs matching the Go engine's tests so plan 008 fixtures can reuse them.

## Done criteria

- [ ] `grep -n "case 'day'\|case 'week'" sdk/tempo/src/calendar/index.ts` shows day/week no longer return fixed ms.
- [ ] TS tempo tests pass, including new DST and locale-parse cases.
- [ ] A `America/New_York` spring-forward `addDays(1)` test asserts the correct local time (matches the Go engine's result for the same input).
- [ ] No Go files modified; no other out-of-scope files modified.
- [ ] `plans/README.md` row for 007 updated.

## STOP conditions

- The strict-vs-lenient behavior for an unmatched month name is a product decision the owner must make (throw vs documented fallback) — implement the default (lenient with an explicit, documented fallback) and note the open question.
- Removing day/week from `fixedUnitMilliseconds` breaks an unrelated consumer of that function — report before widening scope.
- Excerpts don't match live code (drift); a verification fails twice.

## Maintenance notes

- The Go engine is the source of truth for tempo semantics; any future TS arithmetic change should be checked against `pkg/hub/tempo/internal/kernel`. Plan 008 makes this mechanical.
- Reviewer should confirm UTC results are byte-for-byte unchanged (the fix must only affect zoned behavior).

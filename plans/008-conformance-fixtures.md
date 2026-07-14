# Plan 008: Shared Go↔TS conformance fixtures for money and tempo

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/money pkg/hub/tempo sdk/money sdk/tempo`
> Also confirm plans 005, 006, 007 are marked DONE in `plans/README.md`; if not, STOP (this plan depends on the corrected behavior).

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: LOW
- **Depends on**: `plans/005`, `plans/006`, `plans/007`
- **Category**: tests
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

`pkg/hub/money`/`tempo` and `sdk/money`/`tempo` are behavioral twins maintained independently, with **nothing** that mechanically catches divergence. The rounding tie-direction, DST day math, and float-conversion bugs (fixed in 005–007) are all live examples of drift that shipped precisely because no cross-runtime guard exists. This plan adds language-neutral golden fixtures run by both suites, so any future divergence in money rounding or date/time semantics fails CI instead of surfacing as a production discrepancy between backend and frontend.

## Current state

- Twins: `pkg/hub/money` (Go) ↔ `sdk/money` (TS); `pkg/hub/tempo` (Go) ↔ `sdk/tempo` (TS).
- No file in `pkg/hub` references `sdk/*`; no `parity`/`conformance`/`golden` test cross-references the runtimes (verified by grep during the audit).
- Test conventions: Go table-driven tests reading testdata; TS Vitest suites under `sdk/*/tests`. Confirm the repo's test-data conventions by reading an existing Go table test (e.g. `pkg/hub/money/money/manager_test.go`) and a TS test (`sdk/money/tests/…`).

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Go suite | `pnpm exec vp run go:test` | exit 0 |
| TS suite | `pnpm exec vp test` | pass |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: a new shared fixtures directory (JSON), a Go test that loads it, a TS test that loads it, and a short README describing the format. Suggested location: `pkg/hub/testdata/conformance/` or a top-level `conformance/` — pick based on what both runtimes can read without packaging headaches (a repo-root `conformance/` referenced by relative path from both suites is simplest; confirm the TS test runner can read outside its package dir, else place a copy/symlink and document it).

**Out of scope**: changing money/tempo behavior (that is 005–007); adding new features to either package.

## Git workflow

- Branch: `advisor/008-conformance-fixtures`; conventional-commit style.

## Steps

### Step 1: Define the fixture format

Author language-neutral JSON cases: inputs and expected outputs as **strings** (to avoid float/precision ambiguity in JSON). Cover:
- money: `add`/`subtract`/`multiply` (incl. overflow → error marker), `round` at exact halves under the chosen policy, `createFromFloat`, `convert`/`convertWithRate` (incl. a large amount and an overflow marker), `avg`.
- tempo: `addDays`/`addWeeks` across a DST boundary in a named zone, `addMonths` month-end clamping, `diffInMonths`/`Years`, `parseFromPattern` with a non-English locale.

Each case: `{ "op": "...", "args": [...], "expected": "..." | {"error": "overflow"} , "note": "..." }`. Reuse the exact numeric inputs chosen in plans 005–007's tests.

### Step 2: Go loader test

Add a Go test (e.g. `pkg/hub/money/conformance_test.go`, `pkg/hub/tempo/conformance_test.go`) that reads the JSON, dispatches each `op` to the real API, and asserts the string-rendered result equals `expected` (or that the operation returns the marked error). Fail with the case `note` on mismatch.

**Verify**: `cd pkg/hub && go test ./money/... ./tempo/... -run Conformance` → pass.

### Step 3: TS loader test

Add the mirror in `sdk/money/tests` and `sdk/tempo/tests` reading the same JSON and asserting the same expectations.

**Verify**: `pnpm exec vp test` → the conformance suites pass.

### Step 4: Wire into CI signal + document

Ensure both loader tests run under the standard `vp test` / `vp run go:test` so CI catches drift. Add a short `conformance/README.md` (or header comment) explaining: this is the single source of cross-runtime truth; adding a behavior means adding a case here and implementing it in both runtimes.

**Verify**: `pnpm exec vp run go:test && pnpm exec vp test` → both green. `pnpm exec vp run format` → exit 0.

## Test plan

The fixtures **are** the tests. Ensure at least one case per bug fixed in 005–007 so a regression in either runtime re-breaks the conformance suite. Include a deliberately runtime-agnostic edge case (e.g. exact-half rounding) that would have caught the original divergence.

## Done criteria

- [ ] A shared JSON fixture set exists and is read by both a Go and a TS test.
- [ ] `pnpm exec vp run go:test` and `pnpm exec vp test` both pass with the conformance suites active.
- [ ] Each of the 005–007 fixes has at least one conformance case.
- [ ] `conformance/README.md` (or equivalent) documents the format and the "add-a-case" workflow.
- [ ] `plans/README.md` row for 008 updated.

## STOP conditions

- The TS test runner cannot read the chosen fixture path and no clean alternative (copy/symlink with a documented sync) works — report the packaging constraint.
- A conformance case fails because 005–007 did not actually converge the behavior — STOP and report which case/runtime, rather than editing the fixture to match a still-buggy runtime.

## Maintenance notes

- This is the guard that makes the twins maintainable; note in each package's doc that behavior changes require a conformance case.
- Reviewer should confirm the fixtures render outputs as strings (no JSON float precision loss) and that "error" cases are matched by error identity, not message text.

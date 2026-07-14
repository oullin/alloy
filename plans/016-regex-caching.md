# Plan 016: Cache compiled regexes and Intl formatters instead of rebuilding per call

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/str pkg/hub/tempo/parser sdk/tempo`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: LOW
- **Depends on**: none
- **Category**: perf
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

Hot-path primitives recompile regexes (Go) and rebuild `Intl` formatters (TS) on every call:
- **P3**: `str` helpers compile regexes per call — constant patterns (`IsUuid`, `IsUlid`, `Slug`, `Headline`, `Ucsplit`, `Squish`, `Words`, `studlySplit`) and pattern-derived ones (`Is`/glob, `IsMatch`, `Match`/`MatchAll`/`ReplaceMatches`, `Replace`, `Deduplicate`). Consumers call these in loops (slugging lists, validating IDs per record, glob-matching permission names).
- **P4**: `tempo.ParseWithFormat` does `regexp.MustCompile(expression.String())` on every parse, so bulk-parsing a repeated layout pays a full compile per row.
- **P9**: TS tempo builds `Intl.DateTimeFormat`/`NumberFormat`/`RelativeTimeFormat` per call on the `formatIntl`, number, and relative-time paths (the primary `format()` path already caches via `getFormatter`). `Intl.*` constructors are among the most expensive JS date operations.

## Current state

- `pkg/hub/str/str.go`: `Is` (620-624) → `globToRegex` + `regexp.MatchString` (compiles each call); `IsMatch` (650-652); `IsUuid` (701), `IsUlid` (723), `Slug` (1457), `Headline` (444), `Ucsplit` (542), `Squish` (939), `Words` (765), `studlySplit` (387); `Match`/`MatchAll`/`ReplaceMatches` (1125/1145/901), `Replace` (802), `Deduplicate` (952) compile per call.
- `pkg/hub/tempo/parser/format.go:77`: `match := regexp.MustCompile(expression.String()).FindStringSubmatch(input)`. (The fixed patterns in `parser/patterns.go` and `duration/parse.go` are already package-level — good.)
- `sdk/tempo/src/core/index.ts`: `formatIntl` (2322) → `new Intl.DateTimeFormat(...)`; `new Intl.NumberFormat(locale)` (570); `new Intl.RelativeTimeFormat(...)` (2024). `sdk/tempo/src/calendar/index.ts:334,354` also construct `Intl.DateTimeFormat` uncached. The cached pattern to reuse is `getFormatter` at `calendar/index.ts:63-89`.

Convention: Go — hoist constant-pattern regexes to package-level `var` (compiled once at init); cache pattern-derived ones in a `sync.Map` (or mutex-guarded `map[string]*regexp.Regexp`) keyed by the pattern string. TS — a module-level `Map` keyed by `(locale, serialized options, timeZone)`, matching `getFormatter`.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Go tests (str, tempo) | `cd pkg/hub && go test ./str/... ./tempo/...` | exit 0 |
| Go bench | `cd pkg/hub && go test ./str/... -bench . -benchmem` | records improvement |
| Go race (cache concurrency) | `cd pkg/hub && go test -race ./str/...` | exit 0 |
| TS tests | `pnpm exec vp test` (tempo) | pass |
| Full Go suite | `pnpm exec vp run go:test` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `pkg/hub/str/str.go`, `pkg/hub/tempo/parser/format.go`, `sdk/tempo/src/core/index.ts`, `sdk/tempo/src/calendar/index.ts`, and tests/benchmarks. A small compiled-regex cache helper (Go) and formatter caches (TS).

**Out of scope**: `str.go` decomposition (deferred X10); tempo structural refactor; behavior changes of any kind (this is pure caching — outputs must be identical).

## Git workflow

- Branch: `advisor/016-regex-caching`; commit per surface (str, tempo-go, tempo-ts); conventional-commit style.

## Steps

### Step 1: Hoist constant-pattern regexes (Go str)

Move every regex compiled from a *literal* pattern (`IsUuid`, `IsUlid`, `Slug`, `Headline`, `Ucsplit`, `Squish`, `Words`, `studlySplit`, etc.) to package-level `var xRe = regexp.MustCompile(...)`. Replace in-body compilation with the package var.

**Verify**: `cd pkg/hub && go test ./str -run 'Uuid|Ulid|Slug|Squish|Words'` → identical results; bench shows the compile gone.

### Step 2: Cache pattern-derived regexes (Go str)

For `Is`/glob, `IsMatch`, `Match`/`MatchAll`/`ReplaceMatches`, `Replace`, `Deduplicate`, add a compiled-regex cache keyed by the pattern string (`sync.Map` or mutex-guarded map). Compile on miss, reuse on hit. Bound the cache if patterns are unbounded/user-supplied (an LRU or a size cap) to avoid unbounded growth.

**Verify**: `go test -race ./str -run 'Is|Match|Replace'` → identical results, no races; bench shows fewer compiles.

### Step 3: Cache the tempo format regex (Go)

In `parser/format.go`, memoize the compiled regex (and captured token list) keyed by the format pattern string, so repeated parses of the same layout compile once.

**Verify**: `go test ./tempo/parser -run Format` → identical parse results; a benchmark over N parses of one layout shows the compile amortized.

### Step 4: Cache the secondary Intl formatters (TS)

Add module-level `Map` caches keyed by `(locale, serialized options, timeZone)` for the `formatIntl`, `NumberFormat`, and `RelativeTimeFormat` paths (and the two uncached `calendar/index.ts` constructors), reusing the `getFormatter` approach.

**Verify**: `pnpm exec vp test` (tempo) → identical formatting output; (optional) a micro-benchmark or a spy asserting the constructor is called once per distinct key.

### Step 5: Full suites + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp test` → pass; `pnpm exec vp run format` → exit 0.

## Test plan

- Correctness first: existing str/tempo tests must pass unchanged (caching must not alter outputs). Add a concurrency test for the Go pattern cache (`-race`).
- Benchmarks documenting the improvement (str glob/uuid in a loop; tempo bulk parse; TS format in a loop).

## Done criteria

- [ ] Constant-pattern regexes are package-level (`grep -n "regexp.MustCompile\|regexp.Compile" pkg/hub/str/str.go` shows compilation at package scope / cache, not inside per-call hot functions).
- [ ] Pattern-derived regexes are cached; `-race` clean.
- [ ] `tempo/parser/format.go` no longer compiles per parse.
- [ ] TS secondary Intl paths are cached.
- [ ] All str/tempo tests pass unchanged (identical outputs); `pnpm exec vp run go:test` and `pnpm exec vp test` green.
- [ ] No out-of-scope files modified; `plans/README.md` row for 016 updated.

## STOP conditions

- A pattern-derived cache could grow unbounded from user-supplied patterns and no bound is acceptable — report so the owner picks an eviction policy.
- Any cached result differs from the uncached one (a caching bug) — STOP; correctness beats speed.
- Excerpts don't match live code (drift); a test fails after two attempts.

## Maintenance notes

- Note the cache bound/eviction choice so future pattern-heavy callers don't reintroduce unbounded growth.
- Reviewer should confirm outputs are byte-identical to pre-change (this plan must not alter any result).

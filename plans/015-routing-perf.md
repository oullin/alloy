# Plan 015: Index route matching to remove the per-request linear regex scan

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/httpx/routing`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: L
- **Risk**: MED
- **Depends on**: none
- **Category**: perf
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

Every request runs an O(n) scan over all routes for the method, executing a full-path regex per candidate (`matchAgainstRoutes` → `route.Matches` → `uri_validator` `re.MatchString`). A `staticPrefix` is computed at compile time but nothing uses it for indexing. For a consumer with a few hundred routes that is hundreds of regex evaluations per request on the dispatch hot path. Additionally, the compiled ("production") route collection re-filters and allocates a method-filtered slice on every `Get(method)`, and a 404 calls `Get` for all 7 verbs — ~8× full scans plus regex per miss. And `matching.All()` allocates a fresh 4-validator slice per route per request. These are the dispatch-loop costs a route index removes.

## Current state

- `pkg/hub/httpx/routing/route_collection_matching.go:73-97` — `matchAgainstRoutes` iterates every route, calls `route.Matches`.
- `pkg/hub/httpx/routing/matching/uri_validator.go:22-44` — runs the full-path regex per candidate.
- `pkg/hub/httpx/routing/compiler/route_compiler.go:239` — computes `staticPrefix` per route (currently unused by the matcher).
- `pkg/hub/httpx/routing/compiled_route_collection.go:101-119` — `Get(method)` scans all routes and allocates a new `[]*Route` each call; `Match` (94-98) calls it; 404 → `checkForAlternateVerbs` (`route_collection_matching.go:99-116`) calls `Get(m)` for each of 7 verbs.
- `pkg/hub/httpx/routing/matching/validator_interface.go` — `All()` returns a freshly-allocated `[]ValidatorInterface{...}` (4 elements); `route.go:323` calls it inside `Matches` per candidate.
- The dev `RouteCollection` already uses a `map[method][]*Route` — the compiled path regressed relative to it (use it as the reference pattern).

Convention: preserve exact matching semantics — matching order, first-match wins, 405 (method-not-allowed) vs 404, OPTIONS handling. Fully static routes (no `{param}`) are the common case and can go in an exact-match map.

## Commands you will need

| Purpose                  | Command                                                            | Expected                 |
| ------------------------ | ------------------------------------------------------------------ | ------------------------ |
| Routing tests            | `cd pkg/hub && go test ./httpx/routing/...`                        | exit 0                   |
| Benchmark (before/after) | `cd pkg/hub && go test ./httpx/routing/... -bench Match -benchmem` | records ns/op, allocs/op |
| Full Go suite            | `pnpm exec vp run go:test`                                         | exit 0                   |
| Format                   | `pnpm exec vp run format`                                          | exit 0                   |

## Scope

**In scope**: `compiled_route_collection.go`, `route_collection_matching.go`, `matching/validator_interface.go`, and the wiring that builds the compiled collection (to construct indexes once). Tests and a matching benchmark.

**Out of scope**: the route compiler's regex generation (keep it; just index by `staticPrefix`); the dev `RouteCollection` (already fine); router request-scope (plan 004).

## Git workflow

- Branch: `advisor/015-routing-perf`; commit per optimization; conventional-commit style.

## Steps

### Step 0: Establish a baseline benchmark

Add (or extend) a matching benchmark with a realistic route set (e.g. 200 routes, mix of static and dynamic). Record ns/op and allocs/op before any change.

**Verify**: `go test ./httpx/routing -bench Match -benchmem` runs and prints numbers (save them).

### Step 1: Method-indexed compiled collection

Build a `map[string][]*Route` (method → routes) once when the compiled collection is constructed, mirroring the dev `RouteCollection`. Make `Get(method)` an O(1) map read instead of a per-call scan+alloc. Gate `checkForAlternateVerbs` behind a cheap existence check so a 404 doesn't scan all verbs unless needed.

**Verify**: routing tests pass; benchmark shows fewer allocs/op on the 404 and match paths.

### Step 2: Static-path exact-match map + dynamic prefix buckets

Add an exact-match `map[string]*Route` (keyed by method+path) for fully static routes (no params) — these resolve without any regex. Bucket dynamic routes by `staticPrefix`/first segment so only plausible candidates get regex-tested; fall back to the current scan for the residual. Preserve first-match/priority ordering (if two routes could match, the index must return them in the same order the linear scan would have).

**Verify**: routing tests pass (all existing matching cases, including precedence and 405/404); benchmark shows a large ns/op reduction on static-heavy route sets.

### Step 3: Singleton validator set

Make the validator set a package-level singleton slice (the validators are stateless) so `Matches` iterates without allocating per route. Precompute any "without method check" variant.

**Verify**: `go test ./httpx/routing -bench Match -benchmem` shows the per-route validator allocation gone.

### Step 4: Full suite + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

- All existing routing/matching tests must pass unchanged — the index is an optimization, not a semantic change. Add cases that would catch an ordering/precedence regression (two routes where order matters; a static and a dynamic route on overlapping paths; 405 vs 404).
- Benchmark documents the improvement (before/after in the PR description).

## Done criteria

- [ ] All `httpx/routing` tests pass; matching semantics (order, 405/404, OPTIONS) unchanged.
- [ ] `Get(method)` is an O(1) map lookup (no per-call slice allocation) — code review + benchmark allocs/op.
- [ ] Static routes resolve without regex; `staticPrefix` is used by the matcher.
- [ ] Benchmark shows a measurable ns/op and allocs/op improvement vs the step-0 baseline.
- [ ] No out-of-scope files modified; `plans/README.md` row for 015 updated.

## STOP conditions

- The index cannot preserve first-match/priority ordering for overlapping routes without complexity that risks correctness — report; a correct linear scan beats a fast wrong match.
- OPTIONS/405 behavior can't be reproduced through the index — report.
- Excerpts don't match live code (drift); a routing test fails and the fix isn't obvious after two attempts.

## Maintenance notes

- The index is built once at collection-compile time; note that any dynamic route registration after compile must rebuild or update the index.
- Reviewer should focus on ordering/precedence equivalence and the 404/405 paths, not raw speed.

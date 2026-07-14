# Plan 003: Make container resolution concurrency-safe

> **Executor instructions**: Follow step by step. Run every verification command before continuing. On any STOP condition, stop and report. Update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/container`
> On any change, reconcile excerpts against live code; on mismatch, STOP.

## Status

- **Priority**: P1
- **Effort**: L
- **Risk**: MED
- **Depends on**: none
- **Category**: bug
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

`App` documents "All methods are safe for concurrent use" (`container/container.go:35`), but resolution stores per-call state on **container-global** slices and maps and unlocks the mutex while the factory runs. Concurrently: (a) one goroutine's frame on the shared `buildStack` makes another's legitimate resolve look like a circular dependency; (b) `getContextualConcrete` reads `buildStack[len-1]`, so it can apply the wrong goroutine's contextual binding; (c) the LIFO pop removes the wrong frame; (d) `Application`'s deferred-provider maps are read/written with no lock on every `Make`/`Get`, which is a hard `fatal error: concurrent map writes` process crash; (e) singleton resolution can run the factory twice and cache twice. A DI container serving a concurrent HTTP server hits these under normal load. The existing code even comments that `buildStack` is mutated concurrently (`container.go:251`) — it guards the snapshot read but not the underlying logical corruption.

## Current state

- `pkg/hub/container/container.go` — `App` struct (36-58): `mu sync.RWMutex`, plus shared `buildStack []string` and `with []map[string]any`. `resolve` (200-314) locks, clones callback slices, checks the cached instance, unlocks during the factory (275), pushes/pops `buildStack`/`with`. Circular check at 251 (`slices.Contains(c.buildStack, abstract)`); contextual read at 699 (`c.buildStack[len-1]`); parameters via `c.with[len-1]` (342).
- `pkg/hub/container/application.go` — `flushDeferredFor` (97-124) reads/deletes `deferredByKey`, writes `registered[p]`, reads `booted`; called on every `Application.Make`/`MakeWith`/`Get` (129-148). None of the `Application`-level maps are mutex-guarded (only the embedded `*App` has `mu`).

Excerpt (`container.go:265-291`) — shared state mutated across the unlock:
```go
c.buildStack = append(c.buildStack, abstract)
c.with = append(c.with, parameters)
c.mu.Unlock()
...
instance, err := factory(c)
c.mu.Lock()
if len(c.buildStack) > 0 { c.buildStack = c.buildStack[:len(c.buildStack)-1] }
if len(c.with) > 0 { c.with = c.with[:len(c.with)-1] }
c.mu.Unlock()
```

Convention: the container mirrors Laravel's container semantics. The correct fix keeps behavior identical for the single-goroutine case (all existing tests must still pass) while making per-resolution state goroutine-local.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Go tests (container) | `cd pkg/hub && go test ./container/...` | exit 0 |
| Race detector | `cd pkg/hub && go test -race ./container/...` | exit 0, no races |
| Full Go suite | `pnpm exec vp run go:test` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `pkg/hub/container/container.go`, `pkg/hub/container/application.go`, and their test files.

**Out of scope**: the public `App`/`Application` method signatures (keep them stable), binding/tagging/extender semantics beyond making them concurrency-safe.

## Git workflow

- Branch: `advisor/003-container-concurrency`
- Commit per step; conventional-commit style.

## Steps

### Step 1: Make per-resolution state goroutine-local

Replace the container-global `buildStack`/`with` with a resolution context threaded through the resolve call chain. Two acceptable shapes — pick the one that fits the existing code with the least signature churn:
- **(a) Context struct passed down**: introduce an internal `resolution` struct carrying `buildStack []string` and `parameters []map[string]any`, created per top-level `Make`/`resolve` and passed to `getContextualConcrete`, `Parameters()`, and the circular check. Public methods stay unchanged; they construct a fresh `resolution` and pass it inward.
- **(b) Hold the lock for the whole resolution**: keep the state on the container but never unlock during the factory. Simpler, but serializes all resolution (a throughput cost) and risks deadlock if factories re-enter `Make` on the same container — verify re-entrancy before choosing this.

Prefer (a). Whichever you choose, the circular-dependency check, contextual-binding lookup (`container.go:699`), and `Parameters()` (`container.go:342`) must read the goroutine-local state, not shared fields.

**Verify**: existing tests pass (`go test ./container/...`), and a new test resolving two independent graphs from two goroutines does not produce spurious `ErrCircularDependency` or cross-contaminated contextual bindings. `go test -race ./container -run Concurrent` → no races.

### Step 2: Guard singleton building with single-flight

For shared/singleton bindings, ensure concurrent `Make` of the same abstract runs the factory once. Add per-abstract single-flight (e.g. `golang.org/x/sync/singleflight` if already available, or a per-key building lock map). Keep the cached-instance fast path. Guard against deadlock when a factory resolves a *different* abstract (single-flight must be keyed per abstract, not global).

**Verify**: a test where two goroutines `Make` the same singleton whose factory increments a counter asserts the counter is 1 and both receive the same instance. `go test -race ./container -run Singleton` → pass, no races.

### Step 3: Lock the Application-level maps

In `application.go`, add an `Application`-level `sync.Mutex` (separate from the embedded `App.mu`) guarding `deferredByKey`, `registered`, `booted`, and `providers`. Take it in `flushDeferredFor` and any other method mutating those maps. Make `flushDeferredFor` idempotent under the lock (double-checked `registered[p]`).

**Verify**: a test calling `Application.Make` for a deferred-provider key from many goroutines does not panic with "concurrent map writes". `go test -race ./container -run DeferredConcurrent` → pass, no races.

### Step 4: Race-clean the whole container package + full suite

**Verify**: `cd pkg/hub && go test -race ./container/...` → exit 0, no races. `pnpm exec vp run go:test` → exit 0. `pnpm exec vp run format` → exit 0.

## Test plan

New tests in `container/container_test.go` and `application_test.go` (model after existing container tests):
- Concurrent resolution of independent graphs → no false circular errors.
- Concurrent contextual-binding resolution → each goroutine gets its own binding.
- Concurrent singleton resolution → factory runs once.
- Concurrent `Application.Make` on deferred keys → no map-write panic.
All new concurrency tests must pass under `-race`.

## Done criteria

- [ ] `cd pkg/hub && go test -race ./container/...` exits 0 with no race reports.
- [ ] `pnpm exec vp run go:test` exits 0; existing single-goroutine tests unchanged and passing.
- [ ] The "All methods are safe for concurrent use" doc comment (`container.go:35`) is now true; if any method remains explicitly not-safe, the doc says so precisely.
- [ ] No out-of-scope files modified (`git status`).
- [ ] `plans/README.md` row for 003 updated.

## STOP conditions

- Approach (b) is chosen and factories are found to re-enter `Make` on the same container (deadlock risk) — switch to (a) and report.
- `-race` still reports a data race after two fix attempts on the same site.
- Making resolution state local requires changing a public method signature — report before doing so.
- Excerpts don't match live code (drift).

## Maintenance notes

- Reviewer should scrutinize: that single-flight keys per abstract (not globally) to avoid serializing unrelated resolutions, and that contextual bindings still resolve correctly for nested `Make` calls (the parent frame must be visible to the child within the same goroutine).
- Relates to plan 015/P7 (container fast-path lock/allocation cost) — if both land, coordinate the RLock fast-path change so it isn't undone here.

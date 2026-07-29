# Plan 004: Store the "current route" per-request instead of in shared router state

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/httpx/routing/router.go`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: bug
- **Planned at**: commit `bfface5`, 2026-07-14

## RESOLUTION 2026-07-16 (owner decision)

The original plan STOPPED at Step 3: no `Current*` accessor could reach the
request/context without a public signature change, because `MatchableRequest`
exposed no context surface and `runRoute` invoked handlers as zero-arg funcs.

The owner decided to **thread request context through dispatch**; breaking
changes are acceptable pre-GA. Implemented API shape:

1. `contracts.MatchableRequest` gains `Context() context.Context`
   (implementations return their request's context; callers get a
   `context.Background()` fallback when it is nil). In-repo implementers
   updated: `foundation.Request`, the routing test `fakeRequest`.
2. Dispatch derives `ctx = routing.WithCurrentRoute(request.Context(), route)`
   after matching, stores it back on the request (optional `WithContext`
   interface), and invokes handlers with it. `runRoute`'s type switch is
   extended with context-accepting variants —`func(context.Context)`,
   `func(context.Context) error`, `func(context.Context) any`,
   `func(context.Context) (any, error)` — while the zero-arg variants keep
   working unchanged.
3. New package-level context accessors: `routing.CurrentRoute(ctx)`,
   `CurrentRouteName(ctx)`, `CurrentRouteAction(ctx)`, `CurrentRouteIs(ctx, …)`,
   `CurrentRouteNamed(ctx, …)`, `CurrentRouteUses(ctx, …)`. The existing no-arg
   `Router.Current*`/`Is`/`Uses` methods stay but are documented as reading
   process-wide last-matched state, unreliable under concurrent dispatch, and
   **Deprecated** in favor of the ctx accessors.
4. `foundation.RouteResolver` methods take `ctx context.Context` (breaking);
   `foundation.Request` implements `Context()`/`WithContext()` and reads its own
   context in `Fingerprint()`. `handlerx` binds a `routing.ContextRouteResolver`
   (per-request, context-backed) instead of the shared router.

Steps 4–5 of the original plan (removing the shared fields) are intentionally
**not** done: the owner keeps the no-arg methods and their backing fields for
backward compatibility, marked Deprecated.

## Why this matters

The `Router` is a long-lived singleton serving all goroutines, but it stores the matched route for the in-flight request in shared fields `current`/`currentRequest` (`router.go:33-37`), overwritten on every `Dispatch` (`setCurrentRoute`, 738). `Current()`, `CurrentRouteName()`, `Is()`, `Uses()` (634-724) read those fields. The `currentMu` RWMutex prevents a data race but not the logical corruption: under concurrent requests, request B overwrites `current` between request A's dispatch and A's handler calling `Router.Current()`, so A observes B's route. Any middleware or handler that gates behavior on the current route name gets wrong answers under load. (Laravel's router is safe here only because PHP is share-nothing per request.)

## Current state

- `pkg/hub/httpx/routing/router.go`:
    - Fields (33-37): `current *Route`, `currentRequest matching.MatchableRequest`, `currentMu sync.RWMutex`.
    - `DispatchToRoute` (451) calls `r.setCurrentRoute(route)` at line 474.
    - `setCurrentRoute` (738-744): `r.currentMu.Lock(); r.current = route`.
    - Readers: `Current()` (634), `CurrentRouteName()` (655), `CurrentRouteAction()` (666), `Is()` (689), `CurrentRouteNamed` (700), `Uses()` (705), `CurrentRouteUses` (724).

Excerpt:

```go
func (r *Router) setCurrentRoute(route *Route) {
	r.currentMu.Lock()
	defer r.currentMu.Unlock()
	r.current = route
}
```

Convention: requests flow through `matching.MatchableRequest`, which wraps `*http.Request` (read the interface). The idiomatic Go fix is to carry the matched route in the request's `context.Context` and have `Current*` read from the request rather than shared router state.

## Commands you will need

| Purpose            | Command                                           | Expected |
| ------------------ | ------------------------------------------------- | -------- |
| Go tests (routing) | `cd pkg/hub && go test ./httpx/routing/...`       | exit 0   |
| Race detector      | `cd pkg/hub && go test -race ./httpx/routing/...` | exit 0   |
| Full Go suite      | `pnpm exec vp run go:test`                        | exit 0   |
| Format             | `pnpm exec vp run format`                         | exit 0   |

## Scope

**In scope**: `pkg/hub/httpx/routing/router.go` and its tests; a small context-key helper (new unexported type + `context.WithValue`/getter), placed in `router.go` or a sibling `route_context.go`.

**Out of scope**: route matching/compilation; the `Route` type; middleware dispatch ordering beyond threading the context value.

## Git workflow

- Branch: `advisor/004-router-request-scope`
- Commit per step; conventional-commit style.

## Steps

### Step 1: Add a request-scoped current-route context value

Introduce an unexported context key type and helpers:

```go
type currentRouteKey struct{}
func withCurrentRoute(ctx context.Context, route *Route) context.Context { return context.WithValue(ctx, currentRouteKey{}, route) }
func currentRouteFrom(ctx context.Context) *Route { r, _ := ctx.Value(currentRouteKey{}).(*Route); return r }
```

### Step 2: Populate the context at dispatch and thread it into the handler

In `DispatchToRoute` (451), instead of (or in addition to, during migration) `setCurrentRoute`, attach the matched route to the request context and ensure the downstream handler/middleware receive the updated request (`r.WithContext(...)`). Confirm how the request is passed to handlers so the context actually propagates.

### Step 3: Read from the request in the Current* accessors

Change `Current()`, `CurrentRouteName()`, `CurrentRouteAction()`, `Is()`, `Uses()`, `CurrentRouteUses()` to read the route from the request context. If the public signatures don't currently take a request/context, add request-aware variants and have the existing ones delegate — **if you cannot obtain the request in a given accessor without a signature change, STOP and report** so the owner can decide on the API shape.

### Step 4: Remove the shared fields

Once all readers use the context, remove `current`/`currentRequest`/`currentMu` and `setCurrentRoute`. Keep the public method names.

**Verify**: `cd pkg/hub && go test -race ./httpx/routing -run Concurrent` → a test dispatching two requests concurrently and asserting each observes its own route passes with no races.

### Step 5: Full suite + format

**Verify**: `pnpm exec vp run go:test` → exit 0. `pnpm exec vp run format` → exit 0.

## Test plan

- `router_test.go`: existing single-request `Current()`/`Is()`/`Uses()` tests still pass. New concurrency test: dispatch route A on goroutine 1 and route B on goroutine 2, each handler asserts `Current()` returns its own route — must pass under `-race`.

## Done criteria

- [ ] `grep -n "r.current\b\|currentMu\|setCurrentRoute" pkg/hub/httpx/routing/router.go` returns nothing (shared state removed).
- [ ] `cd pkg/hub && go test -race ./httpx/routing/...` exits 0, no races.
- [ ] `pnpm exec vp run go:test` exits 0.
- [ ] Public `Current*`/`Is`/`Uses` method names preserved.
- [ ] No out-of-scope files modified; `plans/README.md` row for 004 updated.

## STOP conditions

- A `Current*` accessor cannot reach the request/context without a public signature change — report for an API decision.
- The request context does not actually propagate to handlers (dispatch copies the request without carrying context) — report, as it indicates a deeper dispatch issue.
- Excerpts don't match live code (drift), or a verification fails twice.

## Maintenance notes

- Reviewer should confirm every `Current*` caller in the codebase (and in the demo app) still compiles and behaves — `grep -rn "\.Current(\|\.CurrentRouteName(\|\.Is(" pkg/hub web/inertia-demo`.
- Future middleware relying on "current route" must read it from the request context, not the router; note this in the router doc comment.

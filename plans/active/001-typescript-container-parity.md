# Plan 001: Ship the TypeScript container and application lifecycle at synchronous Go parity

> **Executor instructions**: Follow this plan step by step. Run every
> verification command and confirm the expected result before moving to the
> next step. If anything in the "STOP conditions" section occurs, stop and
> report — do not improvise. The Improve/Ollin parent maintains plan state;
> do not edit `plans/manifest.json`, `plans/README.md`, plan history, or this
> packet. Do not create worktrees, push, open a pull request, or tag a release.
>
> **Drift check (run first)**: `git diff --stat 49800658e5c6ed36e39163a3004bb5ebef21ec7b..HEAD -- pkg/hub/container pkg/hub/contracts/provider conformance sdk pnpm-workspace.yaml pnpm-lock.yaml package.json vite.config.ts infra web/docs README.md`
> If any in-scope file changed since this plan was written, compare the
> "Current state" facts against the live code. A semantic mismatch is a STOP
> condition.

## Status

- **Priority**: P1
- **Effort**: L
- **Risk**: HIGH
- **Plan ID**: `001-typescript-container-parity`
- **Depends on**: none
- **Category**: migration
- **Planned at**: commit `49800658e5c6ed36e39163a3004bb5ebef21ec7b`, 2026-08-17

## Why this matters

Alloy's service container and provider lifecycle exist only in Go, forcing
TypeScript products to grow local composition roots with overlapping lifetime,
registration, and resolution rules. This plan adds `@hara/sdk-container` as an
idiomatic TypeScript package whose synchronous behavior is guarded against the
tested Go implementation, while adding explicit asynchronous TypeScript APIs
where JavaScript runtimes need them. The package becomes the single reusable
composition primitive that Collections can adopt without moving any product or
Cloudflare policy into Alloy.

## Decisions already made

- The public package is `@hara/sdk-container` under `sdk/container`.
- `ServiceToken<T>` is an opaque runtime object with a stable, non-empty string
  key and generic type marker. `createServiceToken<T>(name)` is the named token
  factory. Equality is by token object identity; `token.name` is for errors and
  introspection, not a second string-based lookup API.
- Bindings accept tokens, not constructors or arbitrary strings. Aliases, tags,
  contextual requirements, provider `provides`, and provider `dependsOn` all
  use typed tokens. Introspection returns tokens and typed metadata.
- Synchronous Go behavior is the parity source of truth. TypeScript exposes
  separate `bind`/`bindAsync`, `make`/`makeAsync`, and sync/async callback,
  extender, call, registration, and boot paths. A sync path that reaches an
  async binding or hook throws `AsyncResolutionRequiredError`; it never leaks a
  Promise as a resolved value.
- Singleton and scoped async construction uses promise singleflight. Concurrent
  callers share the pending promise; success is cached; rejection clears the
  pending entry and is retriable. Scope instances are isolated per child scope.
- Global application helpers are process-oriented compatibility APIs. Package
  documentation must warn that Cloudflare request composition must create an
  application per invocation and must not install it globally.
- The parity policy changes from "container is not a twin candidate" to
  synchronous server-runtime L2 parity. Async APIs remain a documented
  TypeScript extension and do not appear in shared deterministic vectors.
- `ts/v0.3.0` is release preparation only in this pull request. The executor
  must not tag, publish, or claim the release exists. After human merge, the
  owner runs the existing release workflow and verifies the exact asset
  `hara-sdk-container-0.3.0.tgz`.
- The release pack gate requires all direct TypeScript SDK package versions to
  equal `0.3.0`. Updating the four `0.2.1` manifests and retaining
  `navigator-routes` at `0.3.0` is required release-wide metadata, not product
  code churn.

## Current state

- `pkg/hub/container/container.go` defines `App`; the adjacent
  `container_*.go` files implement bindings, resolution parameters, singleton
  and scoped caching, aliases, contextual bindings, tags, extenders, resolving
  callbacks, rebinding callbacks, method bindings, introspection, and flushing.
- `pkg/hub/container/application*.go` wraps `App` with eager/deferred provider
  registration, stable dependency sorting, idempotent boot, late registration,
  reentrant deferred resolution, provider introspection, and two global
  application mechanisms.
- `pkg/hub/contracts/provider/provider.go` declares `Register`, optional
  `Boot`, `Provides`, `Deferred`, and `DependsOn`; dependency keys are service
  abstracts rather than provider classes.
- `pkg/hub/container/errors.go` exposes missing-binding, resolution-cycle,
  alias-cycle, self-alias, missing-method, and missing-global-application
  identities. `RegisterMany` currently panics when the internal stable provider
  sorter reports a dependency cycle. Do not redesign the Go public API merely
  to make TypeScript exceptions look identical.
- `conformance/money.json` and `conformance/tempo.json` show the shared-fixture
  convention. Go and TypeScript tests load the same repository-root JSON.
- `pnpm-workspace.yaml` enumerates each direct SDK. `vite.config.ts` and
  `infra/src/index.ts` enumerate source aliases. TypeScript imports within SDKs
  must use package or package-internal `#` aliases; relative imports fail the
  import gate.
- On the planned base, `console`, `money`, `tempo`, and `workflow` are version
  `0.2.1`; `navigator-routes` is already `0.3.0`. The release workflow packs
  every direct `sdk/*` package and requires a common tag version.
- `web/docs/architecture/parity.md` currently labels container as Go-only and
  explicitly excludes it from twin candidacy. That policy is intentionally
  superseded for this server-runtime package by this approved plan.

## Commands you will need

| Purpose | Command | Expected on success |
| --- | --- | --- |
| Install | `pnpm install --frozen-lockfile` | exit 0 |
| Format | `make format-all` | exit 0 |
| Imports | `pnpm imports:check` | exit 0 |
| Lint | `pnpm exec vp lint` | exit 0 |
| Typecheck | `pnpm exec vp check` | exit 0 |
| TypeScript tests | `pnpm exec vp test --coverage` | exit 0 |
| SDK builds | `pnpm -r --filter './sdk/*' build` | exit 0 |
| Go race/vet suite | `pnpm exec vp run go:test` | exit 0 |
| Documentation | `pnpm docs:build` | exit 0 |
| Packed manifests | `VERSION=0.3.0 bash infra/scripts/tasks/check-pack-ts-packages.sh` | exit 0 and container tarball verified |

Run the repository's `go-coding-standards` and `typescript-coding-standards`
skills before implementation if available. Apply the standards at
`/Users/gocanto/Sites/omniyat-collection/collections/.shared/.agents/skills/go-coding-standards/SKILL.md`
and
`/Users/gocanto/Sites/omniyat-collection/collections/.shared/.agents/skills/typescript-coding-standards/SKILL.md`.

## Scope

**In scope**:

- `sdk/container/**` (new package, source, tests, README, license, package and
  TypeScript build/test configuration)
- `conformance/container.json` and `conformance/README.md`
- `pkg/hub/container/**` only for the shared conformance loader/tests or an
  actual Go defect proven by a failing existing/shared test
- `pkg/hub/contracts/provider/**` only if an atomic Go fix requires it
- `pnpm-workspace.yaml`, `pnpm-lock.yaml`, `package.json`, `vite.config.ts`,
  `infra/src/index.ts`, and the minimum import/pack scripts or tests required to
  register and verify the new package
- `sdk/console/package.json`, `sdk/money/package.json`,
  `sdk/navigator-routes/package.json`, `sdk/tempo/package.json`,
  `sdk/workflow/package.json` for the release-wide `0.3.0` version alignment
- `README.md` and `web/docs/architecture/parity.md`; other `web/docs/**` files
  only when the existing documentation navigation/package index requires an
  entry

**Out of scope**:

- Omniyat/Collections services, Hono, Better Auth, Cloudflare bindings, domain
  adapters, and provider-selection policy
- Changes to Go container behavior without a reproducing test and an atomic
  TypeScript expectation
- Async concepts in `conformance/container.json` or claims that async behavior
  has a Go twin
- A release tag, GitHub release, npm publication, merge, or Collections change
- Browser or Playwright tests

## Git workflow

- Branch: `chore/alloy-typescript-container`, created by the Ollin parent from
  the verified live `origin/main` SHA in a fresh worktree under
  `/Users/gocanto/.codex/worktrees/codex`.
- Commit by phase using the repository's conventional style. Suggested units:
  `feat(container): add synchronous TypeScript container`,
  `feat(container): add async and provider lifecycle APIs`, and
  `chore(release): prepare TypeScript SDK 0.3.0`.
- The executor does not push or open a pull request. The parent independently
  reviews, runs gates, pushes, and opens a draft pull request.

## Steps

### Step 1: scaffold the package and typed public contract

Create `sdk/container` by matching `sdk/workflow`'s package/build/test layout.
Export `ServiceToken<T>`, `createServiceToken<T>`, lifetime/binding/provider
types, `Container`, `Application`, and typed error classes from the package
root. Use package-internal `#container/*` imports and register those aliases in
the package manifest/tsconfig. Add error classes for missing bindings, circular
resolution, alias cycles/self-alias, provider dependency cycles, wrong resolved
types, missing method bindings, missing global application, and sync access to
async work. Errors must preserve token/provider paths as structured readonly
fields and render stable human-readable messages.

**Verify**: `pnpm --filter @hara/sdk-container typecheck` and
`pnpm imports:check` both exit 0.

### Step 2: implement and test the synchronous container

Implement transient, singleton, scoped, and instance bindings; conditional
binding; resolution parameters; aliasing; contextual bindings; tags; sync
extenders; sync resolving/after-resolving callbacks; rebinding callbacks;
method bindings/calls; binding and resolution introspection; instance
forgetting; child scopes; and flushing. Resolution must track a token stack and
throw a typed cycle error with the complete path. Singleton/scoped caches must
store only successfully constructed values. Rebinding or flushing must clear
the affected caches with the same observable semantics as Go.

Use runtime validators on binding factories so a token can optionally carry a
type guard; throw `IncorrectResolvedTypeError` when a guarded token receives an
invalid value. This is the enforceable runtime form of the requested wrong-type
error; do not pretend TypeScript generics alone can validate values.

Mirror every deterministic Go capability in focused TypeScript tests. For API
shape, prefer named options/objects over positional overload ambiguity, while
preserving the Go behavior rather than Go syntax.

**Verify**: `pnpm --filter @hara/sdk-container test -- --coverage` exits 0 and
the package typecheck passes.

### Step 3: add shared synchronous conformance vectors

Add `conformance/container.json` as a restricted declarative operation format.
The fixture may use named scalar/record factories implemented by each test
harness; it must not embed language expressions. Include deterministic cases
for lifetimes, parameters, aliases and alias cycles, contextual selection,
tags, extender/callback order, rebinding invalidation, method binding, deferred
provider triggering, stable provider dependency order, idempotent boot/late
registration, missing binding, and resolution cycles. Each case declares its
operations and expected values/error identity/order/counter totals.

Add a Go test loader under `pkg/hub/container` and a TypeScript loader under
`sdk/container/tests`. Both must execute the same complete case set; neither may
silently skip an unknown operation or vector. Document the fixture schema and
the explicit exclusion of TypeScript-only async behavior in
`conformance/README.md`.

If a vector exposes a real Go defect, first add the failing Go regression test,
then fix Go and TypeScript atomically. If behavior is merely ergonomic or
language-specific, keep Go unchanged and document the TypeScript extension.

**Verify**: the focused Go container tests and TypeScript container tests both
exit 0 and report the same number of shared vectors.

### Step 4: implement asynchronous container extensions

Add explicit async factories/resolution, async extenders/callbacks, async
method calls, and async singleton/scoped caches. `makeAsync` may resolve both
sync and async graphs. `make` must fail immediately when any binding or hook in
the graph is async. Concurrent async singleton/scoped resolution shares one
pending promise; rejected work is removed and a later call retries. Track async
resolution paths across awaits so direct and indirect cycles produce the typed
cycle error rather than hanging.

Add TypeScript-only tests for async factories, mixed graphs, callback/extender
ordering, method calls, singleflight, per-scope isolation, rejection retry,
async cycles, sync/async misuse, rebinding while pending, and flush behavior.
Keep these tests outside the shared JSON.

**Verify**: the container package tests and typecheck both exit 0.

### Step 5: implement Application provider lifecycle and globals

Implement `register`, `registerMany`, `boot`, `registerAsync`,
`registerManyAsync`, and `bootAsync`. Providers declare typed `provides` and
`dependsOn` token lists, plus optional `deferred`; registration sorting is
stable and dependency cycles throw `ProviderCycleError`. Registration and
booting are idempotent by provider object identity. A provider registered after
boot is registered and booted once. Deferred registration triggers on the
first application-level resolution of a provided token, supports reentrant and
concurrent resolution, and boots immediately if the application is already
booted. Sync APIs must reject async provider hooks explicitly.

Add provider introspection for registered, deferred, and booted state. Add the
Go-parity global helpers with explicit set/get/clear behavior and missing-global
errors. Document global helpers as process-only and unsafe for request-scoped
Cloudflare composition.

Test stable sorting, dependency cycles, duplicate/reentrant registration,
deferred flush, boot ordering, late registration, concurrent deferred
singleflight, async rejection retry, provider cycles, sync/async misuse, and
global reset/isolation.

**Verify**: the package tests and typecheck both exit 0.

### Step 6: wire workspace, package checks, docs, and release preparation

Register `sdk/container` in `pnpm-workspace.yaml`, root/source alias tables,
Vite, cache setup only if required by the existing direct-package convention,
the package import check, build graph, and packed-manifest verification. Avoid
special cases where discovery already handles `sdk/*`.

Update the parity matrix to state synchronous server-runtime L2 parity and
TypeScript-only async extensions. Add package usage documentation with sync and
async examples, all lifetimes, provider lifecycle, and the Cloudflare no-global
warning. Update package listings/navigation only where existing siblings are
listed.

Set the new container and all direct TypeScript SDK manifests to `0.3.0`, then
regenerate `pnpm-lock.yaml`. Verify the packed artifact name is exactly
`hara-sdk-container-0.3.0.tgz` and its manifest exports/types contain no source
paths or workspace dependencies.

**Verify**: import check, all SDK builds, documentation build, and
`VERSION=0.3.0 bash infra/scripts/tasks/check-pack-ts-packages.sh` exit 0.

### Step 7: run the complete repository gates and return evidence

Run `make format-all` before the executor's final commit. Then run the commands
in the verification table serially. Inspect `git diff --check`, `git status
--short`, package contents, and the complete diff against the bounded scope.
Return changed files, commit SHAs, test commands/results, shared-vector count,
packed tarball name, and any residual warnings to the parent.

**Verify**: every command exits 0; only in-scope files are modified; no tag,
release, push, or pull request exists from the executor.

## Test plan

- Go and TypeScript loaders execute every vector in
  `conformance/container.json` and fail on unknown operations.
- TypeScript unit suites separately cover the complete sync surface, async
  extensions, Application lifecycle, global helpers, typed error identity, and
  public export/type inference.
- Existing Go tests remain the synchronous behavioral source of truth and run
  under race detection/vet through `go:test`.
- Build/import/pack tests prove consumers can install the packed package in an
  external temporary project and import only declared exports.

## Done criteria

- [ ] `@hara/sdk-container` exposes the requested sync and async container and
  Application APIs with strict types and typed runtime errors.
- [ ] Shared Go/TypeScript deterministic conformance vectors pass with equal
  case counts; async cases remain TypeScript-only and documented.
- [ ] Concurrent async singleton/scoped calls share pending work; rejection is
  retriable; sync access to async work throws the dedicated error.
- [ ] Provider ordering, deferred/reentrant behavior, idempotency, late
  registration, introspection, and global reset are proven by tests.
- [ ] All repository format, imports, lint, typecheck, test, build, Go race/vet,
  docs, and pack gates pass.
- [ ] The packed artifact is `hara-sdk-container-0.3.0.tgz`, but no release is
  tagged or published by this plan.
- [ ] The diff contains only the enumerated Alloy package, parity, conformance,
  documentation, and release-preparation files.
- [ ] Completion evidence is returned to the Improve/Ollin parent.

## STOP conditions

Stop and report without improvising if:

- live `origin/main` no longer contains planned base
  `49800658e5c6ed36e39163a3004bb5ebef21ec7b`, or an in-scope file changed
  semantically after that base;
- an active worktree or open pull request overlaps the in-scope package,
  container, parity, conformance, or release files; stale branches and closed
  pull requests are not active ownership by themselves;
- the Go behavior and an approved requirement conflict and no focused test can
  distinguish a Go defect from an intentional semantic difference;
- implementing wrong-type validation would require pretending erased generic
  types are observable without an explicit token guard;
- release tooling requires product/source changes outside the enumerated
  package manifests and infrastructure files;
- any verification command fails twice after one focused repair;
- credentials, a merge, a tag, or publication are required to proceed.

## Maintenance notes

- Reviewers should scrutinize cache invalidation during pending async work,
  resolution-stack cleanup after rejection, deferred-provider reentrancy, and
  the exact boundary between shared sync parity and TypeScript-only async APIs.
- The post-merge owner action is to create/dispatch `ts/v0.3.0`, verify the
  immutable `hara-sdk-container-0.3.0.tgz` release asset and digest, then unblock
  the Collections adoption packet.
- No Omniyat behavior moves into Alloy. Product provider policy remains in the
  consuming repository.

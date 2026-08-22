# TypeScript HTTP Stack Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Port Alloy's Go application kernel and HTTP stack to TypeScript as two packages — `@hara/sdk-container` (IoC + provider lifecycle) and `@hara/sdk-httpx` (request/response, router, portable middleware, Fetch adapter) — at synchronous server-runtime L2 parity for deterministic behavior.

**Architecture:** There is no `pkg/hub/app` and no top-level `pkg/hub/router`. Application lives in `pkg/hub/container`. Router lives in `pkg/hub/httpx/routing`. TypeScript keeps that map: Application stays inside `@hara/sdk-container`; Router is a concern slice of `@hara/sdk-httpx` exported as `@hara/sdk-httpx/routing`. The HTTP primitive is the Fetch `Request`/`Response` pair, not Node `http` and not Hono. Dispatch threads an `HttpContext` per request; the router never stores the current request as instance state.

**Tech Stack:** TypeScript 7, Node 24+, pnpm 10.33.0, Vite+ / Vitest, Web Fetch API, Web Crypto (HMAC for signed URLs), existing `@hara/sdk-container` (plan 001), shared JSON conformance under `conformance/`.

---

> **Executor instructions**: Follow this plan step by step. Run every
> verification command and confirm the expected result before moving to the
> next step. If anything in the "STOP conditions" section occurs, stop and
> report — do not improvise. The Improve/Ollin parent maintains plan state;
> do not edit `plans/manifest.json`, `plans/README.md`, plan history, or this
> packet. Do not create worktrees, push, open a pull request, or tag a release.
>
> **Drift check (run first)**: `git diff --stat ef933a530b8529840a3ee5dd66219b7aca69c2ae..HEAD -- pkg/hub/container pkg/hub/httpx sdk pnpm-workspace.yaml pnpm-lock.yaml package.json vite.config.ts infra web/docs/architecture/parity.md conformance`
> If any in-scope file changed since this plan was written, compare the
> "Current state" facts against the live code. A semantic mismatch is a STOP
> condition.

## Status

- **Priority**: P1
- **Effort**: XL
- **Risk**: HIGH
- **Plan ID**: `002-typescript-http-stack`
- **Depends on**: `001-typescript-container-parity`
- **Category**: migration
- **Planned at**: commit `ef933a530b8529840a3ee5dd66219b7aca69c2ae`, 2026-08-22

## Why this matters

TypeScript products that want Alloy's bootstrap and HTTP vocabulary currently
reimplement composition, matching, and response shaping next to the Go
originals. This plan adds the missing server-runtime twins so a Node or
Cloudflare worker can `new Application()`, register providers, mount routes,
and return a Fetch `Response` with the same match / 404 / 405 / URL-generation
rules as Go.

## Go to TypeScript map

| The user said | Go source of truth | TypeScript destination |
| --- | --- | --- |
| container | `pkg/hub/container` (`App`) | `@hara/sdk-container` — plan 001 |
| app | `pkg/hub/container` (`Application`) | same package; not `@hara/sdk-app` |
| router | `pkg/hub/httpx/routing` | `@hara/sdk-httpx/routing` |
| httpx | `pkg/hub/httpx/foundation`, `middleware`, `handlerx` | `@hara/sdk-httpx` root + `./middleware` + `./handler` |

Do not create `@hara/sdk-app` or `@hara/sdk-router`. Application has no
separate Go module. Router and foundation share `HttpRequest` / `HttpContext`;
a third package would force a cycle or a shallow types package.

`@hara/sdk-navigator-routes` stays the frontend manifest consumer. The TypeScript
router may emit a `RouteManifest` that package already understands. Do not merge
the packages and do not reimplement `fillPattern` inside httpx.

## Decisions already made

- Plan 001 ships `@hara/sdk-container` first. This plan does not re-specify
  tokens, lifetimes, providers, or async singleflight. If 001 is not merged,
  stop and finish it.
- `@hara/sdk-httpx` is one package with concern-slice exports, matching
  `@hara/sdk-workflow`: `.`, `./routing`, `./foundation`, `./middleware`,
  `./handler`.
- Bindings and route actions use typed tokens and functions. There is no
  `"Handler@method"` string dispatcher and no `reflect`-style injection.
- The portable HTTP types are Fetch `Request` and `Response`.
  `HttpRequest.fromFetch(request)` and `createFetchHandler(router)` are the
  `handlerx` equivalent. Do not depend on Hono, Express, or Node
  `IncomingMessage`.
- Go stores `current` / `currentRequest` on `Router` behind a mutex. TypeScript
  must not. `dispatch` creates an `HttpContext` and passes it through matching,
  middleware, and the action. Cloudflare and concurrent Node requests share one
  router instance.
- GET registration also registers HEAD, as Go does.
- Fallback is a GET/HEAD route on `{fallbackPlaceholder}` with `.*`, marked
  fallback, considered only after ordinary routes miss — same as
  `pkg/hub/httpx/routing/router.go` `Fallback` and
  `route_collection_matching.go` `matchAgainstRoutes`.
- The route compiler emits RE2-safe regex strings. JavaScript `RegExp` executes
  that subset. Shared fixtures must stay in the RE2 ∩ JS intersection (no
  backreferences, no JS-only lookaround). Do not add a RE2 WASM dependency
  unless a vector fails and the parent agrees.
- Resource registration exists as a typed helper
  `resource(name, handlers: ResourceHandlers)`, not PHP controller class names.
  `apiResource` omits `create` and `edit`.
- Signed URLs use HMAC-SHA256 via Web Crypto, same inputs and query names as
  Go `UrlGenerator`.
- Middleware is an onion: `(request, next) => Promise<HttpResponse>`. Aliases,
  groups, and `middlewarePriority` gather in the same order as Go
  `GatherRouteMiddleware`.
- New packages ship at `0.4.0` to match the live SDK manifests. Do not bump
  money, tempo, workflow, console, or navigator-routes in this plan.
- Parity policy: container and httpx become **server-runtime** twins. They are
  not frontend twins. Async dispatch and the Fetch adapter are TypeScript-only
  and stay out of shared fixtures.
- Global `SetApp` remains a process helper from plan 001. Package docs must
  repeat that Cloudflare request composition creates an application per
  invocation and must not install it globally.

## Current state

- `pkg/hub/container` owns `App` and `Application`. Plan 001 is the TypeScript
  twin. On this base, `sdk/container` does not exist yet.
- `pkg/hub/httpx/routing/compiler` is a Symfony-style compiler: `{name}` and
  `{!name}`, optional segments via tokens, `Separators = "/,;.:-_~+*=@|"`,
  `VariableMaximumLength = 32`, RE2 named groups `(?P<name>...)`.
- `pkg/hub/httpx/routing/matching` validates method, URI, host, and scheme.
  URI matching trims a trailing slash (preserving `/`) and path-unescapes
  before the compiled regex.
- `pkg/hub/httpx/routing/route_collection.go` indexes by method, name, and
  action; insertion order is dispatch order; same method+domain+uri overwrites.
- `ErrRouteNotFound` and `MethodNotAllowedError` (with `Allowed` + `Path`) are
  the match miss identities. OPTIONS can synthesize an Allow route.
- `pkg/hub/httpx/foundation` wraps `*http.Request` / `http.ResponseWriter`.
  TypeScript wraps Fetch instead; behavior (input merge, JSON, redirect,
  no-content, status/headers/cookies) is the parity target, not the Go types.
- `pkg/hub/httpx/handlerx` is the net/http adapter. TypeScript replacement is
  `createFetchHandler`.
- `pkg/hub/httpx/middleware` has recovery, CORS, cache headers, frame guard,
  trust hosts, trust proxies, path encoding, post size, request log, check
  response, preload assets. v1 ports the first eight; skip Redis throttle and
  preload-assets.
- `pkg/hub/httpx/server` is a `net/http.Server` runner. Out of scope.
- `@hara/sdk-navigator-routes` fills `{param}`, `{param?}`, `{param:regex}`,
  and `:param`. It does not match or dispatch.
- Live SDK versions are `0.4.0`. Plan 001's `0.3.0` release note is stale;
  do not relitigate it here.
- `web/docs/architecture/parity.md` still lists `container` and `httpx` as
  Go-only and not twin candidates. This plan supersedes that for server
  runtimes.

## Commands you will need

| Purpose | Command | Expected on success |
| --- | --- | --- |
| Install | `pnpm install --frozen-lockfile` | exit 0 |
| Format | `pnpm exec vp run format-all` | exit 0 |
| Imports | `pnpm imports:check` | exit 0 |
| Lint | `pnpm lint` | exit 0 |
| Typecheck | `pnpm exec vp check` | exit 0 |
| TypeScript tests | `pnpm exec vp test` | exit 0 |
| Focused httpx | `pnpm --filter @hara/sdk-httpx test` | exit 0 |
| Focused httpx types | `pnpm --filter @hara/sdk-httpx typecheck` | exit 0 |
| SDK builds | `pnpm -r --filter './sdk/*' build` | exit 0 |
| Go race/vet | `pnpm exec vp run go:test` | exit 0 |
| Docs | `pnpm docs:build` | exit 0 |

Run `@go-coding-standards` only if a Go conformance loader needs a fix.
Run `@typescript-coding-standards` and `@no-relative-module-specifiers`
before writing TypeScript. Use `#httpx/*` inside the package; never `./` or
`../` specifiers.

Copy package/build/test layout from `sdk/workflow`: `package.json`,
`tsconfig.json`, `tsconfig.build.json`, `tsconfig.package.json`, `LICENSE.md`,
`README.md`, `src/index.ts`, `tests/*.test.ts`. Tests import from
`@hara/sdk-httpx` and `@hara/sdk-httpx/routing`, not from `#httpx/*`.

## Scope

**In scope**:

- `sdk/httpx/**` (new package)
- `conformance/routing.json` and a routing section in `conformance/README.md`
- Go loaders under `pkg/hub/httpx/routing` only for shared vectors or a
  proven Go defect
- Workspace registration: `pnpm-workspace.yaml`, `pnpm-lock.yaml`,
  `tsconfig.json`, `vite.config.ts`, `infra/src/index.ts`
- `README.md`, `web/docs/architecture/parity.md`, and navigation/package
  index entries that already list siblings
- A `RoutingServiceProvider` that binds the router on an Application from
  `@hara/sdk-container`

**Out of scope**:

- Re-implementing plan 001 (container, Application, tokens, async singleflight)
- `@hara/sdk-app`, `@hara/sdk-router`, Hono, Express, Fastify
- `pkg/hub/httpx/server` (net/http server)
- Redis throttle, SQS, session, cookie, CSRF, Inertia, precognition
- Filesystem-backed uploads (export a `FileStore` port only)
- PHP/Go `"Handler@method"` and `reflect` action injection
- Process-global current route on `Router`
- Merging or rewriting `@hara/sdk-navigator-routes`
- Omniyat/Collections product policy
- A release tag, publish, merge, or pull request from the executor

## Git workflow

- Branch: `chore/alloy-typescript-http-stack`, created by the parent from
  live `origin/main` after plan 001 has landed or is the merge base.
- Commit by task using the repository conventional style. Suggested units:
  `feat(httpx): scaffold TypeScript HTTP package`,
  `feat(httpx): add route compiler and matching`,
  `feat(httpx): add router dispatch and URL generation`,
  `feat(httpx): add Fetch request, response, and handler`,
  `feat(httpx): add portable middleware and routing provider`.
- The executor does not push or open a pull request.

---

### Task 1: Confirm plan 001 is available

**Files:** none (read-only)

**Step 1: Check that `@hara/sdk-container` exists**

Run:

```sh
test -f sdk/container/package.json && node -p "require('./sdk/container/package.json').name"
```

Expected: `@hara/sdk-container`

If the file is missing, stop. This plan depends on 001. Do not scaffold a
second container.

**Step 2: Confirm the public symbols this package will import**

Read `sdk/container/src/index.ts`. You need `Application`, `createServiceToken`,
and a `Make`-compatible surface for `BindingContainer`:

```ts
type BindingContainer = {
	make(token: ServiceToken<unknown>): unknown;
};
```

If those names differ, use the live 001 exports. Do not invent a parallel
container type.

**Step 3: Commit**

No commit. Proceed.

---

### Task 2: Scaffold `@hara/sdk-httpx`

**Files:**

- Create: `sdk/httpx/package.json`
- Create: `sdk/httpx/tsconfig.json`
- Create: `sdk/httpx/tsconfig.build.json`
- Create: `sdk/httpx/tsconfig.package.json`
- Create: `sdk/httpx/LICENSE.md` (copy `sdk/workflow/LICENSE.md`)
- Create: `sdk/httpx/README.md`
- Create: `sdk/httpx/src/index.ts`
- Create: `sdk/httpx/src/errors.ts`
- Create: `sdk/httpx/src/routing/index.ts`
- Create: `sdk/httpx/src/foundation/index.ts`
- Create: `sdk/httpx/src/middleware/index.ts`
- Create: `sdk/httpx/src/handler/index.ts`
- Create: `sdk/httpx/tests/exports.test.ts`
- Modify: `pnpm-workspace.yaml` — add `"sdk/httpx"`
- Modify: `tsconfig.json` — add `{ "path": "./sdk/httpx" }`
- Modify: `vite.config.ts` — add `@hara/sdk-httpx` and `#httpx/*` aliases
- Modify: `infra/src/index.ts` — add `'@hara/sdk-httpx': resolve(sdkRoot, 'httpx', 'src')`

**Step 1: Write the failing export test**

```ts
import { describe, expect, it } from 'vite-plus/test';

describe('package exports', () => {
	it('exposes the root error identities', async () => {
		const httpx = await import('@hara/sdk-httpx');

		expect(httpx.RouteNotFoundError).toBeTypeOf('function');
		expect(httpx.MethodNotAllowedError).toBeTypeOf('function');
		expect(httpx.HttpResponseError).toBeTypeOf('function');
	});

	it('exposes routing from the concern subpath', async () => {
		const routing = await import('@hara/sdk-httpx/routing');

		expect(routing.HTTP_VERBS).toEqual(['GET', 'HEAD', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS']);
	});
});
```

**Step 2: Run test to verify it fails**

Run: `pnpm --filter @hara/sdk-httpx test`

Expected: FAIL — package not in the workspace / module not found.

**Step 3: Write the package manifest and empty barrels**

`sdk/httpx/package.json`:

```json
{
	"name": "@hara/sdk-httpx",
	"version": "0.4.0",
	"private": true,
	"publishConfig": {
		"access": "restricted",
		"registry": "https://hara.sh/npm/"
	},
	"type": "module",
	"types": "./dist/index.d.ts",
	"main": "./dist/index.js",
	"module": "./dist/index.js",
	"license": "MIT",
	"repository": {
		"type": "git",
		"url": "git+https://github.com/oullin/alloy.git",
		"directory": "sdk/httpx"
	},
	"sideEffects": false,
	"files": ["dist"],
	"exports": {
		".": {
			"types": "./dist/index.d.ts",
			"import": "./dist/index.js",
			"default": "./dist/index.js"
		},
		"./routing": {
			"types": "./dist/routing/index.d.ts",
			"import": "./dist/routing/index.js",
			"default": "./dist/routing/index.js"
		},
		"./foundation": {
			"types": "./dist/foundation/index.d.ts",
			"import": "./dist/foundation/index.js",
			"default": "./dist/foundation/index.js"
		},
		"./middleware": {
			"types": "./dist/middleware/index.d.ts",
			"import": "./dist/middleware/index.js",
			"default": "./dist/middleware/index.js"
		},
		"./handler": {
			"types": "./dist/handler/index.d.ts",
			"import": "./dist/handler/index.js",
			"default": "./dist/handler/index.js"
		}
	},
	"imports": {
		"#httpx/*": {
			"types": "./dist/*.d.ts",
			"default": "./dist/*.js"
		}
	},
	"engines": {
		"node": ">=22"
	},
	"scripts": {
		"build": "bash -lc 'source ../../infra/scripts/tasks/cache-env.sh && rm -rf dist \"${ALLOY_CACHE_PATH}/tsbuild/httpx-build.tsbuildinfo\" && tsc -b tsconfig.build.json --force'",
		"build:package": "rm -rf dist && tsc -p tsconfig.package.json",
		"prepare": "pnpm build:package",
		"prepack": "pnpm build",
		"test": "bash -lc 'source ../../infra/scripts/tasks/cache-env.sh && vp test run tests'",
		"typecheck": "bash -lc 'source ../../infra/scripts/tasks/cache-env.sh && tsc -p tsconfig.json --noEmit'"
	},
	"dependencies": {
		"@hara/sdk-container": "workspace:*"
	},
	"devDependencies": {
		"@types/node": "^26.2.0",
		"typescript": "^7.0.2"
	}
}
```

Mirror `sdk/workflow` tsconfigs. Paths:

```json
"@hara/sdk-httpx": ["./src/index.ts"],
"@hara/sdk-httpx/*": ["./src/*"],
"#httpx/*": ["./src/*"]
```

`src/errors.ts` — tagged errors with structured fields:

```ts
/** No route, including any other HTTP verb, responds to the request. */
export class RouteNotFoundError extends Error {
	readonly _tag = 'RouteNotFoundError';

	constructor(readonly path: string) {
		super(`route not found: ${path}`);
		this.name = 'RouteNotFoundError';
	}
}

/** A route exists at the URI but not for the requested method. */
export class MethodNotAllowedError extends Error {
	readonly _tag = 'MethodNotAllowedError';

	constructor(
		readonly path: string,
		readonly allowed: readonly string[],
	) {
		super(`the requested method is not supported for route ${path}; supported methods: ${allowed.join(', ')}`);
		this.name = 'MethodNotAllowedError';
	}
}

/** An error that already knows the HTTP status to render. */
export class HttpResponseError extends Error {
	readonly _tag = 'HttpResponseError';

	constructor(
		readonly statusCode: number,
		message: string,
		readonly headers: Readonly<Record<string, string>> = {},
	) {
		super(`httpx: HTTP ${statusCode}: ${message}`);
		this.name = 'HttpResponseError';
	}
}

export const HTTP_VERBS = ['GET', 'HEAD', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'] as const;
```

Vite aliases (append next to the workflow block):

```ts
{
	find: '@hara/sdk-httpx',
	replacement: repoPath('./sdk/httpx/src'),
},
{
	find: /^@hara\/sdk-httpx\/(.+)$/u,
	replacement: repoPath('./sdk/httpx/src/$1'),
},
{
	find: /^#httpx\/(.+)$/u,
	replacement: repoPath('./sdk/httpx/src/$1'),
},
```

**Step 4: Run test to verify it passes**

```sh
pnpm install
pnpm --filter @hara/sdk-httpx typecheck
pnpm --filter @hara/sdk-httpx test
pnpm imports:check
```

Expected: PASS.

**Step 5: Commit**

```bash
git add sdk/httpx pnpm-workspace.yaml pnpm-lock.yaml tsconfig.json vite.config.ts infra/src/index.ts
git commit -m "$(cat <<'EOF'
feat(httpx): scaffold the TypeScript HTTP package

EOF
)"
```

---

### Task 3: Route compiler

**Files:**

- Create: `sdk/httpx/src/routing/compiler/token.ts`
- Create: `sdk/httpx/src/routing/compiler/compiled-route.ts`
- Create: `sdk/httpx/src/routing/compiler/route-compiler.ts`
- Create: `sdk/httpx/tests/compiler.test.ts`
- Test: later also `conformance/routing.json`

**Step 1: Write the failing compiler tests**

Port the deterministic cases from `pkg/hub/httpx/routing/compiler` tests. Minimum:

```ts
import { describe, expect, it } from 'vite-plus/test';

import { RouteCompiler } from '@hara/sdk-httpx/routing';

describe('RouteCompiler', () => {
	const compiler = new RouteCompiler();

	it('compiles a static path', () => {
		const compiled = compiler.compile({ path: '/users', host: '', defaults: {}, requirements: {} });

		expect(compiled.staticPrefix).toBe('/users');
		expect(compiled.variables).toEqual([]);
		expect(compiled.compiledRegex.test('/users')).toBe(true);
		expect(compiled.compiledRegex.test('/users/1')).toBe(false);
	});

	it('compiles a required variable', () => {
		const compiled = compiler.compile({ path: '/users/{id}', host: '', defaults: {}, requirements: {} });

		expect(compiled.variables).toEqual(['id']);
		expect(compiled.staticPrefix).toBe('/users/');
		expect(compiled.compiledRegex.test('/users/42')).toBe(true);
	});

	it('applies a where requirement', () => {
		const compiled = compiler.compile({
			path: '/users/{id}',
			host: '',
			defaults: {},
			requirements: { id: '[0-9]+' },
		});

		expect(compiled.compiledRegex.test('/users/42')).toBe(true);
		expect(compiled.compiledRegex.test('/users/abc')).toBe(false);
	});

	it('rejects _fragment as a path parameter', () => {
		expect(() => compiler.compile({ path: '/x/{_fragment}', host: '', defaults: {}, requirements: {} })).toThrow(/_fragment/);
	});

	it('rejects names longer than 32 characters', () => {
		const name = 'a'.repeat(33);

		expect(() => compiler.compile({ path: `/{${name}}`, host: '', defaults: {}, requirements: {} })).toThrow(/32/);
	});
});
```

**Step 2: Run test to verify it fails**

Run: `pnpm --filter @hara/sdk-httpx test`

Expected: FAIL — `RouteCompiler` is not exported.

**Step 3: Write the compiler**

Read `pkg/hub/httpx/routing/compiler/route_compiler.go` and
`pkg/hub/httpx/routing/contracts/compiler/compiled_route.go` in full before
typing. Port `compilePattern` token-for-token:

- `SEPARATORS = '/,;.:-_~+*=@|'`
- `VARIABLE_MAXIMUM_LENGTH = 32`
- Variable regex: `\{(!)?([\w\x80-\xff]+)\}`
- Named groups must be `[A-Za-z_][A-Za-z0-9_]*`
- Important `!` variables are never optional
- Optional tokens use the default map, same as Go
- Host pattern compiles first, then path; variables merge
- JS `RegExp` source is the Go regex with `(?P<name>...)` rewritten to
  `(?<name>...)` — that rewrite is TypeScript-only and must be covered by a
  unit test, not a shared fixture (the fixture stores the logical tokens and
  match outcomes, not the engine-specific group syntax)

`CompiledRoute` is an immutable DTO: private constructor, static `from(...)`,
readonly getters for `staticPrefix`, `regex`, `tokens`, `pathVariables`,
`hostRegex`, `hostTokens`, `hostVariables`, `variables`, plus
`compiledRegex` / `compiledHostRegex` as `RegExp | undefined`.

**Step 4: Run test to verify it passes**

Run: `pnpm --filter @hara/sdk-httpx test`

Expected: PASS. If a Go test exists that you cannot explain, add it to the
TypeScript suite before changing Go.

**Step 5: Commit**

```bash
git add sdk/httpx
git commit -m "$(cat <<'EOF'
feat(httpx): port the Symfony-style route compiler

EOF
)"
```

---

### Task 4: Shared routing conformance vectors

**Files:**

- Create: `conformance/routing.json`
- Modify: `conformance/README.md`
- Create: `pkg/hub/httpx/routing/conformance_test.go`
- Create: `sdk/httpx/tests/conformance.test.ts`

**Step 1: Write the fixture**

Language-neutral operations only. No JS or Go expressions.

```json
[
	{
		"op": "compile",
		"path": "/users/{id}",
		"host": "",
		"defaults": {},
		"requirements": { "id": "[0-9]+" },
		"expected": {
			"staticPrefix": "/users/",
			"variables": ["id"],
			"pathVariables": ["id"]
		},
		"note": "required numeric id"
	},
	{
		"op": "match",
		"routes": [{ "methods": ["GET", "HEAD"], "uri": "/users/{id}", "requirements": { "id": "[0-9]+" } }],
		"request": { "method": "GET", "path": "/users/42", "host": "example.test", "secure": false },
		"expected": { "uri": "/users/{id}", "parameters": { "id": "42" } },
		"note": "GET match plus parameter bind"
	},
	{
		"op": "match",
		"routes": [{ "methods": ["GET", "HEAD"], "uri": "/users" }],
		"request": { "method": "POST", "path": "/users", "host": "example.test", "secure": false },
		"error": "ERR_METHOD_NOT_ALLOWED",
		"expectedAllowed": ["GET", "HEAD"],
		"note": "405 lists the verbs that do match the URI"
	},
	{
		"op": "match",
		"routes": [{ "methods": ["GET", "HEAD"], "uri": "/users" }],
		"request": { "method": "GET", "path": "/missing", "host": "example.test", "secure": false },
		"error": "ERR_ROUTE_NOT_FOUND",
		"note": "unknown path is 404, not 405"
	},
	{
		"op": "match",
		"routes": [
			{ "methods": ["GET", "HEAD"], "uri": "/users/{id}" },
			{ "methods": ["GET", "HEAD"], "uri": "{fallbackPlaceholder}", "wheres": { "fallbackPlaceholder": ".*" }, "fallback": true }
		],
		"request": { "method": "GET", "path": "/nope", "host": "example.test", "secure": false },
		"expected": { "fallback": true },
		"note": "fallback runs only after ordinary routes miss"
	},
	{
		"op": "generate",
		"uri": "/users/{id}",
		"params": { "id": "42" },
		"expected": "/users/42",
		"note": "token fill from compiled path"
	}
]
```

Add cases for: trailing-slash trim, path unescape, host constraint, https-only
(`secure`), group prefix, name lookup, overwrite of same method+domain+uri,
insertion-order match, optional segment defaults. Keep async and Fetch out.

Document in `conformance/README.md`: both loaders fail on an unknown `op`;
error identity is `ERR_ROUTE_NOT_FOUND` / `ERR_METHOD_NOT_ALLOWED` /
`ERR_COMPILE`; TypeScript-only Fetch adapter tests stay in `sdk/httpx/tests`.

**Step 2: Write failing loaders**

Go loader walks every object, fails on unknown `op`, and compares expected
fields. TypeScript loader does the same. Neither may skip.

**Step 3: Implement enough matching/generation for the vectors**

If compile works but match/generate do not yet, keep those vectors in the
file and implement Tasks 5–7 until both loaders report the same case count.

**Step 4: Verify**

```sh
pnpm exec vp run go:test
pnpm --filter @hara/sdk-httpx test
```

Expected: both loaders execute the same number of vectors; unknown ops fail.

**Step 5: Commit**

```bash
git add conformance/routing.json conformance/README.md pkg/hub/httpx/routing/conformance_test.go sdk/httpx
git commit -m "$(cat <<'EOF'
test(httpx): add shared routing conformance vectors

EOF
)"
```

---

### Task 5: Route, collection, and validators

**Files:**

- Create: `sdk/httpx/src/routing/route.ts`
- Create: `sdk/httpx/src/routing/route-collection.ts`
- Create: `sdk/httpx/src/routing/matching/method-validator.ts`
- Create: `sdk/httpx/src/routing/matching/uri-validator.ts`
- Create: `sdk/httpx/src/routing/matching/host-validator.ts`
- Create: `sdk/httpx/src/routing/matching/scheme-validator.ts`
- Create: `sdk/httpx/tests/route-collection.test.ts`

**Step 1: Write failing collection tests**

```ts
import { describe, expect, it } from 'vite-plus/test';

import { MethodNotAllowedError, Route, RouteCollection, RouteNotFoundError } from '@hara/sdk-httpx/routing';

const request = (method: string, path: string) => ({
	method,
	host: 'example.test',
	pathInfo: path,
	secure: false,
});

describe('RouteCollection.match', () => {
	it('matches the first registered route in insertion order', () => {
		const routes = new RouteCollection();
		const first = new Route(['GET', 'HEAD'], '/users/{id}', () => 'one');
		const second = new Route(['GET', 'HEAD'], '/users/{id}', () => 'two');

		routes.add(first);
		routes.add(second);

		const matched = routes.match(request('GET', '/users/1'));

		expect(matched.action).toBe(second.action);
	});

	it('throws MethodNotAllowedError with the allowed verbs', () => {
		const routes = new RouteCollection();
		routes.add(new Route(['GET', 'HEAD'], '/users', () => 'ok'));

		expect(() => routes.match(request('POST', '/users'))).toThrow(MethodNotAllowedError);
	});

	it('throws RouteNotFoundError when nothing matches', () => {
		expect(() => new RouteCollection().match(request('GET', '/missing'))).toThrow(RouteNotFoundError);
	});
});
```

**Step 2: Run test to verify it fails**

Expected: FAIL — `RouteCollection` missing.

**Step 3: Implement**

`Route` is an orchestrator: uri, methods, action function, defaults, wheres,
name, domain, https-only, fallback flag, compiled cache. `compile()` calls
`RouteCompiler`. `whereAlpha` / `whereNumber` / `whereUuid` copy the Go
regex helpers.

`RouteCollection.add` uses the same keys as Go: per-verb
`domain + uri`, and `methods.join('|') + domain + uri` for `allRoutes`.
Name lookup keeps the first registered name (Go `addLookups`).

Validators implement:

```ts
type MatchableRequest = {
	method(): string;
	host(): string;
	pathInfo(): string;
	secure(): boolean;
};

type Validator = {
	matches(route: Route, request: MatchableRequest): boolean;
};
```

`UriValidator`: trim trailing `/` except root, `decodeURIComponent` the path,
then `compiledRegex.test(path)`.

Matching order: non-fallback routes for the verb, then fallbacks, then
alternate-verb scan for 405, then 404.

**Step 4: Run tests**

Expected: PASS, including the `match` vectors in `conformance/routing.json`.

**Step 5: Commit**

```bash
git add sdk/httpx
git commit -m "$(cat <<'EOF'
feat(httpx): add route collection matching and 404/405 identities

EOF
)"
```

---

### Task 6: Router registration API

**Files:**

- Create: `sdk/httpx/src/routing/router.ts`
- Create: `sdk/httpx/src/routing/resource.ts`
- Create: `sdk/httpx/tests/router-register.test.ts`

**Step 1: Write failing registration tests**

```ts
import { describe, expect, it } from 'vite-plus/test';

import { Router } from '@hara/sdk-httpx/routing';

describe('Router registration', () => {
	it('registers GET and HEAD together', () => {
		const router = new Router();
		router.get('/users', () => 'ok');

		expect(router.routes().all()[0]?.methods()).toEqual(['GET', 'HEAD']);
	});

	it('prefixes grouped uris and names', () => {
		const router = new Router();

		router.group({ prefix: '/admin', as: 'admin.' }, (group) => {
			group.get('/users', () => 'ok').name('users');
		});

		expect(router.routes().getByName('admin.users')?.uri).toBe('/admin/users');
	});

	it('registers the REST resource verbs', () => {
		const router = new Router();

		router.resource('users', {
			index: () => 'index',
			create: () => 'create',
			store: () => 'store',
			show: () => 'show',
			edit: () => 'edit',
			update: () => 'update',
			destroy: () => 'destroy',
		});

		expect(router.routes().getByName('users.index')?.methods()).toEqual(['GET', 'HEAD']);
		expect(router.routes().getByName('users.store')?.methods()).toEqual(['POST']);
		expect(router.routes().getByName('users.update')?.methods()).toEqual(['PUT', 'PATCH']);
		expect(router.routes().getByName('users.destroy')?.methods()).toEqual(['DELETE']);
	});

	it('omits create and edit on apiResource', () => {
		const router = new Router();
		router.apiResource('users', { index: () => 'i', store: () => 's', show: () => 'g', update: () => 'u', destroy: () => 'd' });

		expect(router.routes().getByName('users.create')).toBeUndefined();
		expect(router.routes().getByName('users.edit')).toBeUndefined();
	});
});
```

**Step 2: Run test to verify it fails**

Expected: FAIL — `Router` missing.

**Step 3: Implement**

`Router` methods: `get`, `post`, `put`, `patch`, `delete`, `options`, `any`,
`match`, `fallback`, `redirect`, `permanentRedirect`, `group`, `resource`,
`apiResource`, `pattern` / `patterns`, `aliasMiddleware`, `middlewareGroup`,
`name` via the returned `Route`.

Group attributes are a typed object, not `map[string]any`:

```ts
type RouteGroupAttributes = {
	readonly prefix?: string;
	readonly as?: string;
	readonly domain?: string;
	readonly middleware?: readonly MiddlewareReference[];
	readonly where?: Readonly<Record<string, string>>;
	readonly https?: boolean;
};
```

`resource` URIs match Go/Laravel: `name`, `name/create`, `name`,
`name/{name}`, `name/{name}/edit`, `name/{name}`, `name/{name}`. Nested
resources and singleton resources can wait until the Go tests you port
require them; do not invent extra resource flavors.

Actions are:

```ts
type RouteAction = (context: HttpContext) => HttpResponse | Response | Promise<HttpResponse | Response>;
```

No string class names.

**Step 4: Run tests**

Expected: PASS.

**Step 5: Commit**

```bash
git add sdk/httpx
git commit -m "$(cat <<'EOF'
feat(httpx): add typed router registration, groups, and resources

EOF
)"
```

---

### Task 7: Dispatch, middleware gather, and URL generation

**Files:**

- Create: `sdk/httpx/src/routing/http-context.ts`
- Create: `sdk/httpx/src/routing/middleware-gather.ts`
- Create: `sdk/httpx/src/routing/url-generator.ts`
- Create: `sdk/httpx/tests/router-dispatch.test.ts`
- Create: `sdk/httpx/tests/url-generator.test.ts`

**Step 1: Write failing dispatch tests**

```ts
import { describe, expect, it } from 'vite-plus/test';

import { HttpRequest } from '@hara/sdk-httpx/foundation';
import { Router } from '@hara/sdk-httpx/routing';

describe('Router.dispatch', () => {
	it('runs route middleware outside the action and not as router state', async () => {
		const order: string[] = [];
		const router = new Router();

		router.aliasMiddleware('log', async (request, next) => {
			order.push('before');
			const response = await next(request);
			order.push('after');
			return response;
		});

		router.get('/ping', (ctx) => {
			order.push(`action:${ctx.request.pathInfo()}`);
			return ctx.response.json({ ok: true });
		}).middleware('log');

		const first = HttpRequest.fromFetch(new Request('https://example.test/ping'));
		const second = HttpRequest.fromFetch(new Request('https://example.test/ping'));

		await Promise.all([router.dispatch(first), router.dispatch(second)]);

		expect(order.filter((step) => step === 'action:/ping')).toHaveLength(2);
		expect(router.current()).toBeUndefined();
	});

	it('does not leak the matched route across concurrent dispatches', async () => {
		const router = new Router();
		router.get('/a', (ctx) => ctx.response.text(ctx.route?.name() ?? ''));
		router.get('/b', (ctx) => ctx.response.text(ctx.route?.name() ?? ''));
		router.routes().getByName; // names set below
	});
});
```

Name the routes `a` and `b` in the implementation of that last test and
assert each response body equals its own name when the two dispatches run
concurrently.

URL generation tests: `to`, `route`, `signedRoute`, `hasValidSignature`,
`forceHttps`, `forceRootUrl`. Port vectors from
`pkg/hub/httpx/routing/url_generator_test.go` that do not need a session.

**Step 2: Run tests to verify they fail**

Expected: FAIL — `dispatch` missing.

**Step 3: Implement**

```ts
class HttpContext {
	constructor(
		readonly request: HttpRequest,
		readonly response: HttpResponseFactory,
		readonly route: Route | undefined,
		readonly container: BindingContainer | undefined,
	) {}
}

class Router {
	async dispatch(request: MatchableRequest): Promise<DispatchResult> {
		const route = this.routes().match(request);
		const context = new HttpContext(HttpRequest.fromMatchable(request), new HttpResponseFactory(), route, this.container);
		const pipeline = this.gatherRouteMiddleware(route);
		const value = await this.runPipeline(pipeline, context, () => route.invoke(context));

		return { route, value, context };
	}
}
```

`current()` returns `undefined` on the TypeScript router. Document this as an
intentional divergence from Go. Tests must not add a process-wide slot later
"for parity."

Signed URLs: HMAC-SHA256 over the canonical URL + expiry query keys Go uses.
Read `url_generator.go` and copy the query parameter names. Use
`crypto.subtle` (Web Crypto). Tests supply a fixed key.

Emit `toManifest(): RouteManifest` using the type from
`@hara/sdk-navigator-routes` so a frontend can consume the same names. Add a
test that `fillPattern(manifest[name], params)` equals `urlGenerator.route(name, params)`
for brace patterns without host/scheme. Colon-only navigator patterns are
out of scope for the Go compiler.

**Step 4: Run tests**

Expected: PASS, including `generate` conformance vectors.

**Step 5: Commit**

```bash
git add sdk/httpx
git commit -m "$(cat <<'EOF'
feat(httpx): dispatch through request-scoped context and generate URLs

EOF
)"
```

---

### Task 8: Foundation request and response

**Files:**

- Create: `sdk/httpx/src/foundation/http-request.ts`
- Create: `sdk/httpx/src/foundation/http-response.ts`
- Create: `sdk/httpx/src/foundation/http-response-factory.ts`
- Create: `sdk/httpx/tests/request.test.ts`
- Create: `sdk/httpx/tests/response.test.ts`

**Step 1: Write failing request/response tests**

Port behavior from `request_test.go`, `request_input_test.go`,
`response_test.go`, `response_json_test.go`, `response_redirect_test.go`:

- method, path, host, scheme, query, header accessors
- `all()` merges query then body (JSON object or form)
- `only` / `except` / `input` / `boolean` / `integer`
- `wantsJson` / `accepts` / `expectsJson` from Accept and X-Requested-With
- `json()`, `text()`, `html()`, `noContent()`, `redirect()`, `cookie()`
- cookies buffer until `toFetch()`
- `HttpResponseError` renders status + headers

Do not port session flash, old input, or precognition in v1. Keep
`SessionStore` and `RouteResolver` as ports on `HttpRequest` so later
packages can attach them.

**Step 2: Run tests to verify they fail**

Expected: FAIL.

**Step 3: Implement**

`HttpRequest` is an orchestrator around a Fetch `Request`. Cache parsed input
on first `all()`. `fromFetch` is the constructor path. `pathInfo()` is the
URL pathname. `secure()` is `url.protocol === 'https:'`.

`HttpResponse` is an immutable DTO built by `HttpResponseFactory` (fluent
orchestrator). `toFetch()` returns a Fetch `Response`. A raw Fetch `Response`
returned from an action passes through the handler unchanged.

**Step 4: Run tests**

Expected: PASS.

**Step 5: Commit**

```bash
git add sdk/httpx
git commit -m "$(cat <<'EOF'
feat(httpx): wrap Fetch Request and Response with Alloy accessors

EOF
)"
```

---

### Task 9: Fetch handler

**Files:**

- Create: `sdk/httpx/src/handler/fetch-handler.ts`
- Create: `sdk/httpx/tests/fetch-handler.test.ts`

**Step 1: Write the failing adapter tests**

```ts
import { describe, expect, it } from 'vite-plus/test';

import { createFetchHandler } from '@hara/sdk-httpx/handler';
import { Router } from '@hara/sdk-httpx/routing';

describe('createFetchHandler', () => {
	it('returns JSON from a matched action', async () => {
		const router = new Router();
		router.get('/ping', (ctx) => ctx.response.json({ ok: true }));

		const response = await createFetchHandler(router)(new Request('https://example.test/ping'));

		expect(response.status).toBe(200);
		await expect(response.json()).resolves.toEqual({ ok: true });
	});

	it('renders 404 and 405 with Allow', async () => {
		const router = new Router();
		router.get('/only-get', () => new Response('ok'));
		const handle = createFetchHandler(router);

		const missing = await handle(new Request('https://example.test/nope'));
		expect(missing.status).toBe(404);

		const wrong = await handle(new Request('https://example.test/only-get', { method: 'POST' }));
		expect(wrong.status).toBe(405);
		expect(wrong.headers.get('Allow')).toContain('GET');
	});
});
```

**Step 2: Run test to verify it fails**

Expected: FAIL.

**Step 3: Implement**

Mirror `pkg/hub/httpx/handlerx/handler.go` `writeError` / `writeResult`:

- `RouteNotFoundError` → 404
- `MethodNotAllowedError` → 405 + `Allow`
- `HttpResponseError` → its status, headers, message
- `HttpResponse` → `toFetch()`
- Fetch `Response` → as-is
- `string` → `text/html` 200 (Go writes strings as the body)
- `undefined` / `null` → 204
- unknown → 500, no stack in the body

**Step 4: Run tests**

Expected: PASS.

**Step 5: Commit**

```bash
git add sdk/httpx
git commit -m "$(cat <<'EOF'
feat(httpx): add the Fetch request handler adapter

EOF
)"
```

---

### Task 10: Portable middleware

**Files:**

- Create: `sdk/httpx/src/middleware/recovery.ts`
- Create: `sdk/httpx/src/middleware/cors.ts`
- Create: `sdk/httpx/src/middleware/cache-headers.ts`
- Create: `sdk/httpx/src/middleware/frame-guard.ts`
- Create: `sdk/httpx/src/middleware/trust-hosts.ts`
- Create: `sdk/httpx/src/middleware/trust-proxies.ts`
- Create: `sdk/httpx/src/middleware/validate-path-encoding.ts`
- Create: `sdk/httpx/src/middleware/validate-post-size.ts`
- Create: `sdk/httpx/src/middleware/request-log.ts`
- Create: `sdk/httpx/tests/middleware.test.ts`

**Step 1: Write failing tests from the Go `*_test.go` files**

One describe per middleware. Recovery converts thrown errors into
`HttpResponseError` / 500 and does not swallow `HttpResponseError`. CORS
reflects configured origins. Frame guard sets `X-Frame-Options`. Cache
headers set the documented Cache-Control. Path encoding rejects `%2f` /
malformed escapes the same way as Go. Post size rejects bodies over the
limit with `ErrPostTooLarge`. Trust hosts / proxies only accept the
configured patterns and rewrite scheme/host from `X-Forwarded-*` when
trusted.

Request log takes an injected `Logger` port; the test uses an in-memory
logger. No `console.log` in the middleware itself.

**Step 2: Run tests to verify they fail**

Expected: FAIL.

**Step 3: Implement as classes**

```ts
export class RecoveryMiddleware {
	constructor(private readonly report?: (error: unknown) => void) {}

	async handle(request: HttpRequest, next: Next): Promise<HttpResponse> {
		try {
			return await next(request);
		} catch (error) {
			this.report?.(error);
			return HttpResponse.fromError(error);
		}
	}
}
```

Skip `preload_assets` and Redis throttle.

**Step 4: Run tests**

Expected: PASS.

**Step 5: Commit**

```bash
git add sdk/httpx
git commit -m "$(cat <<'EOF'
feat(httpx): port portable HTTP middleware to Fetch

EOF
)"
```

---

### Task 11: Routing provider and docs

**Files:**

- Create: `sdk/httpx/src/routing/routing-service-provider.ts`
- Create: `sdk/httpx/tests/provider.test.ts`
- Modify: `sdk/httpx/README.md`
- Modify: `README.md`
- Modify: `web/docs/architecture/parity.md`
- Modify: `web/docs/.vuepress/config.ts` only if sibling SDK packages are listed
- Modify: `web/docs/.vuepress/theme/components/home/data.ts` only if sibling SDKs appear there

**Step 1: Write the failing provider test**

```ts
import { describe, expect, it } from 'vite-plus/test';

import { Application, createServiceToken } from '@hara/sdk-container';
import { Router, RoutingServiceProvider, routerToken } from '@hara/sdk-httpx/routing';

describe('RoutingServiceProvider', () => {
	it('binds a router and boots idempotently', () => {
		const app = new Application();
		app.register(new RoutingServiceProvider());
		app.boot();

		const router = app.make(routerToken);

		expect(router).toBeInstanceOf(Router);
		app.boot();
		expect(app.make(routerToken)).toBe(router);
	});
});
```

If plan 001 uses different register/boot names, follow those.

**Step 2: Implement the provider**

`provides` includes `routerToken`. `register` binds a singleton `Router`.
`boot` is empty in v1 (Go attaches events; this plan has a noop dispatcher).

README must show:

```ts
const app = new Application();
app.register(new RoutingServiceProvider());
app.boot();

const router = app.make(routerToken);
router.get('/ping', (ctx) => ctx.response.json({ ok: true }));

export default { fetch: createFetchHandler(router) };
```

and the Cloudflare warning: create the application per isolate/invocation;
do not call `setApp` on the request path.

Parity matrix: `container` and `httpx` become **Both (twin)**, level **L2**
for the shared routing fixtures, **L1** for request/response wrappers.
Record TypeScript-only items (Fetch adapter, async actions, no
`Router.current`) in the divergence register with stable IDs (`X14` onward).

**Step 3: Verify**

```sh
pnpm --filter @hara/sdk-httpx test
pnpm --filter @hara/sdk-httpx typecheck
pnpm docs:build
```

Expected: PASS.

**Step 4: Commit**

```bash
git add sdk/httpx README.md web/docs
git commit -m "$(cat <<'EOF'
feat(httpx): bind the router through Application and document server-runtime parity

EOF
)"
```

---

### Task 12: Repository gates

**Files:** only those already touched.

**Step 1: Format**

```sh
pnpm exec vp run format-all
```

**Step 2: Run gates serially**

```sh
pnpm imports:check
pnpm lint
pnpm exec vp check
pnpm exec vp test
pnpm -r --filter './sdk/*' build
pnpm exec vp run go:test
pnpm docs:build
```

Expected: every command exits 0.

**Step 3: Inspect the diff**

```sh
git diff --check
git status --short
```

Expected: only in-scope files; no `sdk/app`, no Hono, no session/Inertia.

**Step 4: Commit format leftovers if fmtkit changed files**

New commit, do not amend.

**Step 5: Return evidence**

Changed files, commit SHAs, shared-vector count, commands and results. No
tag, push, or pull request.

---

## Test plan

- Go and TypeScript loaders execute every vector in `conformance/routing.json`
  and fail on unknown operations.
- TypeScript unit suites cover compiler edge cases, collection overwrite and
  order, 404/405, groups, resources, concurrent dispatch isolation, signed
  URLs, input merge, response factories, Fetch adapter mapping, and each
  portable middleware.
- A navigator-routes interop test fills a brace pattern from `toManifest()`
  and matches `UrlGenerator.route`.
- Existing Go httpx tests remain the matching/compiler source of truth and
  run under `go:test`.
- `pnpm --filter @hara/sdk-httpx build` plus the workspace import gate prove
  consumers import only declared exports.

## Done criteria

- [ ] Plan 001 `@hara/sdk-container` is importable and used by the routing provider
- [ ] `@hara/sdk-httpx` exposes routing, foundation, middleware, and handler subpaths
- [ ] Shared routing vectors pass with equal case counts in Go and TypeScript
- [ ] Concurrent `dispatch` calls do not share `HttpContext` or current-route state
- [ ] `createFetchHandler` maps 404, 405, JSON, and `HttpResponseError`
- [ ] Portable middleware listed in Task 10 is tested
- [ ] Parity docs describe server-runtime twins and the `Router.current` divergence
- [ ] Format, imports, lint, typecheck, test, build, Go race/vet, and docs gates pass
- [ ] Diff stays inside the enumerated scope

## STOP conditions

Stop and report without improvising if:

- plan 001 is absent or its public API cannot bind a singleton router;
- live `origin/main` no longer contains planned base
  `ef933a530b8529840a3ee5dd66219b7aca69c2ae`, or an in-scope file changed
  semantically after that base;
- a shared vector and Go disagree and no focused test can tell a Go defect
  from an intentional language difference;
- RE2 and JavaScript `RegExp` disagree on a fixture and the only fix is a
  new WASM dependency;
- implementing `"Handler@method"` or `Router.current` looks necessary to
  make a Go test pass — keep the TypeScript API and document the divergence;
- any verification command fails twice after one focused repair;
- credentials, a merge, a tag, or publication are required to proceed.

## Maintenance notes

- Reviewers should scrutinize compiler token optionalization, 405 allowed
  verb order, concurrent dispatch, and the Fetch error-mapping table.
- Do not treat Hono compatibility as follow-up inside Alloy. Consumers may
  wrap `createFetchHandler` themselves.
- Session, cookie, and CSRF stay Go-only until those packages are planned
  as their own twins.

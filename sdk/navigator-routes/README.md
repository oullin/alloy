# Navigator Routes TypeScript

The TypeScript package lives in `sdk/navigator-routes` and exposes
`@hara/sdk-navigator-routes`. It is a tiny, dependency-free resolver that turns
named routes into URLs from a manifest of route patterns. Parameter values are
URL-encoded, duplicate slashes are collapsed, and — unless strict mode is on —
unknown route names fall back to a configurable sentinel URL instead of
throwing.

This is a private workspace package: it is consumed by sibling packages via
`workspace:*` and is never published to npm.

## Pattern syntax

Two syntaxes are supported and can be mixed inside a single pattern:

| Token            | Meaning                                                          |
| ---------------- | ---------------------------------------------------------------- |
| `{param}`        | required, Laravel-style                                          |
| `{param?}`       | optional; the whole `/{param?}` segment is dropped when unfilled |
| `{param:regex}`  | required with a binding constraint; the constraint is ignored    |
| `{param:regex?}` | optional with a binding constraint                               |
| `:param`         | required, Express/Hono-style; the name is `[A-Za-z0-9_]+`        |

Because a `:param` name cannot contain `/`, a protocol colon never matches:
`https://cdn.example.com/assets/:asset` fills only `:asset`. Brace tokens are
matched before colon tokens, so the `:slug` inside `{post:slug}` is never
mistaken for a parameter of its own.

```ts
import { fillPattern } from '@hara/sdk-navigator-routes';

fillPattern(
	'/users/:user/posts/:post',
	{ user: 7, post: 'hello world' },
);
// "/users/7/posts/hello%20world"
fillPattern(
	'/teams/:team/members/{member}/{tab?}',
	{ team: 'core', member: 3 },
);
// "/teams/core/members/3"
```

## Resolving from a manifest

Build a resolver once from a route manifest and call it by name:

```ts
import { createRouteResolver } from '@hara/sdk-navigator-routes';

const route = createRouteResolver(
	{
		'posts.show': '/posts/{post}',
		'docs.page': '/docs/{page?}',
	},
);

route(
	'posts.show',
	{ post: 42 },
).url; // "/posts/42"
route('docs.page').url; // "/docs"
route('missing').url; // "#!expose:unknown-route" (and warns via console)
```

The manifest can also be a function, so routes injected at runtime (for
example from a server-rendered page) are re-read on every call:

```ts
import { createRouteResolver } from '@hara/sdk-navigator-routes';

const route = createRouteResolver(
	() => window.__routes,
	{
		fallback: '/',
		onMissingRoute: (name) => reportMissingRoute(name),
	},
);
```

The lower-level helpers are exported for one-off use:

```ts
import { fillPattern, resolveRoute, resolveRouteUrl } from '@hara/sdk-navigator-routes';

fillPattern(
	'/users/{user}/posts/{post?}',
	{ user: 7 },
); // "/users/7/posts"
resolveRouteUrl(
	manifest,
	'posts.show',
	{ post: 42 },
); // "/posts/42"
resolveRoute(
	manifest,
	'posts.show',
	{ post: 42 },
); // { url: "/posts/42" }
```

## Typed pattern parameters

`PatternParams<TPattern>` reads the parameter names straight out of a string
literal pattern: `{param}` and `:param` become required keys, `{param?}`
becomes an optional key, and a `{param:regex}` constraint contributes `param`.

```ts
import type { PatternParams } from '@hara/sdk-navigator-routes';

type Params = PatternParams<'/teams/:team/members/{member}/{tab?}'>;
// { team: RouteParamValue; member: RouteParamValue; tab?: RouteParamValue }
```

`fillPattern` is generic over its pattern, so a literal pattern types — and
requires — its own parameters:

```ts
fillPattern(
	'/teams/:team/members/{member}',
	{ team: 'core', member: 3 },
); // ok
fillPattern(
	'/teams/:team/members/{member}',
	{ team: 'core' },
); // error: `member` is missing
fillPattern(
	'/teams/:team',
	{ team: 'core', slug: 'x' },
); // error: `slug` is not a parameter
```

A pattern whose type is a widened `string` (the manifest case) keeps the old
behaviour exactly: the parameters argument is the untyped `RouteParams` record
and is optional. `null` and `undefined` patterns behave as before too.

## Strict mode

`strict: true` removes the two pieces of guesswork this package was built with:

- `fillPattern` throws — naming the pattern and the key — when a **required**
  placeholder is left unfilled, instead of passing the raw `{param}` or
  `:param` text through into the URL. Optional `{param?}` segments still
  collapse silently.
- `resolveRouteUrl`, `resolveRoute`, and resolvers from `createRouteResolver`
  throw — naming the route — when a route name is missing from the manifest,
  instead of returning the `#!expose:unknown-route` sentinel. The
  `onMissingRoute` reporter is not called; the throw is the report.
- `fillPattern(null)` / `fillPattern(undefined)` throw instead of returning the
  fallback URL.

```ts
const route = createRouteResolver(
	manifest,
	{ strict: true },
);

route(
	'posts.show',
	{ post: 42 },
).url; // "/posts/42"
route('typo.here'); // throws: [navigator-routes] unknown route "typo.here"
fillPattern(
	'/posts/{post}',
	{},
	{ strict: true },
); // throws, names `post`
```

**Recommended: `strict: true` for new code.** The default is non-strict purely
for backwards compatibility. The legacy behaviour — returning the
`#!expose:unknown-route` magic string for an unknown route, and silently
leaving an unfilled `{param}` in the URL — turns a routing mistake into a
broken link that only shows up in production. Strict mode turns both into an
exception at the call site. Nothing changes for existing callers who do not
opt in.

## API overview

| Export                                               | Purpose                                                                |
| ---------------------------------------------------- | ---------------------------------------------------------------------- |
| `createRouteResolver(manifest, options?)`            | builds a `RouteResolver` bound to a manifest (object or lazy function) |
| `resolveRoute(manifest, name, params?, options?)`    | resolves a named route to a `RouteResult` (`{ url }`)                  |
| `resolveRouteUrl(manifest, name, params?, options?)` | same as `resolveRoute`, returning the URL string directly              |
| `fillPattern(pattern, params?, options?)`            | substitutes params into a single route pattern                         |
| `PatternParams<TPattern>`                            | the parameters a literal pattern names                                 |
| `FillPatternArgs<TPattern>`                          | the trailing `fillPattern` arguments for a given pattern               |
| `RouteManifest`, `RouteParams`, `RouteParamValue`    | manifest and parameter types                                           |
| `RouteResult`, `RouteResolver`                       | resolver result and resolver types                                     |
| `NavigatorOptions`, `MissingRouteReporter`           | fallback URL, strict mode, and missing-route reporting hooks           |

Outside strict mode, unresolved required parameters are left as-is in the URL
(e.g. `/posts/{post}`), while unresolved optional segments are dropped. Tests
run with `pnpm --filter @hara/sdk-navigator-routes test`.

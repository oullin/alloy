# Navigator Routes TypeScript

The TypeScript package lives in `sdk/navigator-routes` and exposes
`@alloy/sdk/navigator-routes`. It is a tiny, dependency-free resolver that turns
named routes into URLs from a manifest of Laravel-style patterns. Patterns
support required (`{id}`) and optional (`{page?}`) parameters with optional
constraints (`{id:\d+}`); parameter values are URL-encoded, duplicate
slashes are collapsed, and unknown route names fall back to a configurable
sentinel URL instead of throwing.

This is a private workspace package: it is consumed by sibling packages via
`workspace:*` and is never published to npm.

Build a resolver once from a route manifest and call it by name:

```ts
import { createRouteResolver } from '@alloy/sdk/navigator-routes';

const route = createRouteResolver({
	'posts.show': '/posts/{post}',
	'docs.page': '/docs/{page?}',
});

route('posts.show', { post: 42 }).url; // "/posts/42"
route('docs.page').url; // "/docs"
route('missing').url; // "#!expose:unknown-route" (and warns via console)
```

The manifest can also be a function, so routes injected at runtime (for
example from a server-rendered page) are re-read on every call:

```ts
import { createRouteResolver } from '@alloy/sdk/navigator-routes';

const route = createRouteResolver(() => window.__routes, {
	fallback: '/',
	onMissingRoute: (name) => reportMissingRoute(name),
});
```

The lower-level helpers are exported for one-off use:

```ts
import { fillPattern, resolveRoute, resolveRouteUrl } from '@alloy/sdk/navigator-routes';

fillPattern('/users/{user}/posts/{post?}', { user: 7 }); // "/users/7/posts"
resolveRouteUrl(manifest, 'posts.show', { post: 42 }); // "/posts/42"
resolveRoute(manifest, 'posts.show', { post: 42 }); // { url: "/posts/42" }
```

## API overview

| Export | Purpose |
| --- | --- |
| `createRouteResolver(manifest, options?)` | builds a `RouteResolver` bound to a manifest (object or lazy function) |
| `resolveRoute(manifest, name, params?, options?)` | resolves a named route to a `RouteResult` (`{ url }`) |
| `resolveRouteUrl(manifest, name, params?, options?)` | same as `resolveRoute`, returning the URL string directly |
| `fillPattern(pattern, params?, options?)` | substitutes params into a single route pattern |
| `RouteManifest`, `RouteParams`, `RouteResult`, `RouteResolver` | manifest and resolver types |
| `NavigatorOptions`, `MissingRouteReporter` | fallback URL and missing-route reporting hooks |

Unresolved required parameters are left as-is in the URL (e.g.
`/posts/{post}`), while unresolved optional segments are dropped. Tests run
with `pnpm --filter @alloy/sdk/navigator-routes test`.

# @hara/sdk-httpx

Portable TypeScript HTTP stack providing Symfony-style routing, Fetch Request/Response foundation primitives, middleware pipelines, and Fetch handler adapters with cross-runtime Go parity.

## Features

- **Router**: Tree-like route registration, Symfony-style route compilation and matching, URL generation, HMAC-SHA256 signed URLs.
- **Foundation**: `HttpRequest` and `HttpResponse` wrappers around standard Fetch API `Request` and `Response`.
- **Middleware**: Portable middleware pipeline for CORS, cache headers, frame guards, trust hosts, trust proxies, and recovery.
- **Handler**: `createFetchHandler` adapter connecting router and middleware to web standard Fetch handlers (Cloudflare Workers, Deno, Bun, Node `undici`).
- **Container Integration**: `RoutingServiceProvider` for `@hara/sdk-container` integration.

## Usage

```ts
import { Application } from '@hara/sdk-container';
import { createFetchHandler } from '@hara/sdk-httpx/handler';
import { routerToken, RoutingServiceProvider } from '@hara/sdk-httpx/routing';

const app = new Application();

app.register(new RoutingServiceProvider());
app.boot();

const router = app.make(routerToken);

router.get('/ping', (ctx) => ctx.response.json({ ok: true }));

router.get('/users/{id}', (ctx) => {
	const id = ctx.route?.parameter('id');

	return ctx.response.json({ user_id: id });
});

export default {
	fetch: createFetchHandler(router),
};
```

### Worker / Serverless Lifecycle Note

In serverless or edge environments (such as Cloudflare Workers), instantiate the `Application` and `Router` once per isolate or per invocation according to your lifecycle strategy. Do not mutate shared global state during request execution.

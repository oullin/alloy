# @hara/sdk-container

`@hara/sdk-container` is Alloy's typed service container and application
provider lifecycle for TypeScript server runtimes. Its synchronous behavior is
an L2-conformant mirror of `hara.sh/alloy/container`; explicit async APIs are a
TypeScript-only extension.

```ts
import { Container, createServiceToken } from '@hara/sdk-container';

const logger = createServiceToken<{ info(message: string): void }>('logger');
const container = new Container();

container.singleton(logger, () => ({ info: console.info }));
container.make(logger).info('ready');
```

## Lifetimes and async work

`bind` is transient by default. Use `singleton` for a container-wide cached
value, `scoped` for a child-scope cached value, or `instance` for a value that
already exists. `bindAsync`, `singletonAsync`, and `scopedAsync` are explicit:
call `makeAsync` for graphs that use them. Calling `make` on async work throws
`AsyncResolutionRequiredError` rather than returning a promise as a service.

```ts
const connection = createServiceToken<{ close(): Promise<void> }>('connection');

container.singletonAsync(connection, async () => await openConnection());

const db = await container.makeAsync(connection);
```

Concurrent `makeAsync` calls share singleton and scoped construction. A rejected
construction is not cached and can be retried. Each concurrent caller still runs
its own resolving/after callback snapshot after the shared value is ready.

Synchronous `make` preflights participating hooks before factories or caches run:
async bindings, before/resolving/after callbacks, and extenders reject with
`AsyncResolutionRequiredError` and leave no factory, cache, or resolved mutation.
Cache hits only preflight before-resolving callbacks, matching Go.

`rebinding` and `refresh` are synchronous Go-parity APIs. There is no async
rebinding/refresh subscription in this release — a later sync `bind`/`instance`
must not mutate and then fail mid-notification. Async resolving callbacks,
extenders, and methods remain supported TypeScript extensions.

`when(...tokens)` installs the same contextual binding for each concrete token.
For configuration, use the typed `giveConfig(configToken, getter, fallback?)`
form rather than a string lookup: it resolves `configToken`, passes the typed
value to `getter`, and uses `fallback` when configuration is unavailable or the
getter returns `undefined`. Without a fallback, an unavailable configuration
still raises its original resolution error.

`ServiceToken<T>` is opaque and invariant; use `AnyServiceToken` for
heterogeneous metadata. Tokens with the same diagnostic name remain separate
services. Runtime token guards may validate values, including `null` and
`undefined`, which are valid cached values when the guard permits them.

## Providers

`Application` adds stable provider dependency ordering, deferred registration,
idempotent booting, and late-registration booting. `booted()` is true inside
sync and async boot hooks; a failed initial `bootAsync()` resets application
boot state so concurrent singleflight waiters reject together and a later retry
can succeed. Registered-but-unbooted providers remain retryable through
`registerAsync(provider)` and `bootAsync()` even after an earlier successful
application boot. Calling sync `boot()` while async boot work is pending throws
`AsyncResolutionRequiredError` without changing boot state.

```ts
import { Application, createServiceToken } from '@hara/sdk-container';

const settings = createServiceToken<{ readonly region: string }>('settings');
const app = new Application();

app.register({
	provides: [settings],
	register(application) {
		application.instance(settings, { region: 'ap-southeast-1' });
	},
});
app.boot();
```

Global application helpers are process-oriented compatibility APIs only. Do not
use them in Cloudflare request composition: create an `Application` per request
and pass it through the request composition root instead.

The Go-compatible helpers are `getInstance`/`setInstance` for a plain process
container and `setApp`, `global`, `hasApp`, `make`, and `resolve` for an
application. The `must*` spellings are throwing aliases. `childScope()` and all
async APIs are TypeScript extensions rather than shared conformance behavior.
Calling `childScope()` on an `Application` returns another `Application`: its
scoped values are isolated while provider, deferred-registration, and boot
lifecycle state remain shared.

Tests run with `pnpm --filter @hara/sdk-container test`.

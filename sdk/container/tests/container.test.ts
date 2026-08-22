import { describe, expect, it } from 'vite-plus/test';

import {
	AliasCycleError,
	Application,
	AsyncResolutionRequiredError,
	CircularResolutionError,
	Container,
	IncorrectResolvedTypeError,
	MissingBindingError,
	MissingGlobalApplicationError,
	ProviderCycleError,
	MissingMethodBindingError,
	clearGlobalApplication,
	createServiceToken,
	getInstance,
	getGlobalApplication,
	make,
	setInstance,
	setGlobalApplication,
	type AnyServiceProvider,
	type AnyServiceToken,
	type AsyncServiceProvider,
	type ServiceToken,
} from '@hara/sdk-container';

type Assert<T extends true> = T;

type IsAssignable<From, To> = From extends To ? true : false;

const tokenIsInvariant: Assert<IsAssignable<ServiceToken<number>, ServiceToken<string>> extends false ? true : false> = true;

// @ts-expect-error - exported erased tokens are nominal too, not structural objects.
const forgedAnyToken: AnyServiceToken = { name: 'forged' };
// @ts-expect-error - an async provider must explicitly implement async registration.
const incompleteAsyncProvider: AsyncServiceProvider = {};
// @ts-expect-error - rebindingAsync is not part of the public container surface.
const removedRebindingAsync: Container['rebindingAsync'] = undefined;
// @ts-expect-error - refreshAsync is not part of the public container surface.
const removedRefreshAsync: Container['refreshAsync'] = undefined;

void forgedAnyToken;
void incompleteAsyncProvider;
void tokenIsInvariant;
void removedRebindingAsync;
void removedRefreshAsync;

describe('@hara/sdk-container', () => {
	it('resolves lifetimes, parameters, aliases, tags, and contextual bindings', () => {
		const container = new Container();
		const transient = createServiceToken<number>('transient');
		const singleton = createServiceToken<{ id: number }>('singleton');
		const scoped = createServiceToken<{ id: number }>('scoped');
		const consumer = createServiceToken<string>('consumer');

		let next = 0;

		container.bind(transient, (_container, parameters) => Number(parameters.value ?? 1));
		container.singleton(singleton, () => ({ id: ++next }));
		container.scoped(scoped, () => ({ id: ++next }));
		container.alias(singleton, createServiceToken<{ id: number }>('singleton-alias'));
		container.tag([transient, singleton], 'values');
		container.when(consumer).needs(transient).give(7);
		container.bind(consumer, (resolved) => `value:${resolved.make(transient)}`);

		expect(container.makeWith(transient, { value: 4 })).toBe(4);
		expect(container.make(singleton)).toBe(container.make(singleton));
		expect(container.make(scoped)).toBe(container.make(scoped));
		expect(container.childScope().make(scoped)).not.toBe(container.make(scoped));
		expect(container.make(consumer)).toBe('value:7');
		expect(container.tagged('values')).toHaveLength(2);
	});

	it('orders extenders and callbacks, invalidates rebinding caches, and validates guards', () => {
		const container = new Container();

		const token = createServiceToken<number>('number', { guard: (value): value is number => typeof value === 'number' });

		const events: string[] = [];

		container.bind(token, () => 1, { lifetime: 'singleton' });
		container.beforeResolving(token, () => events.push('before'));
		container.extend(token, (value) => value + 1);
		container.resolving(token, (value) => events.push(`resolving:${value}`));
		container.afterResolving(token, (value) => events.push(`after:${value}`));
		expect(container.make(token)).toBe(2);
		expect(events).toEqual(['before', 'resolving:2', 'after:2']);
		container.bind(token, () => 4, { lifetime: 'singleton' });
		expect(container.make(token)).toBe(5);
		container.instance(token, 9);
		expect(container.make(token)).toBe(9);
		expect(() => container.instance(token, 'not-a-number' as never)).toThrow(IncorrectResolvedTypeError);
	});

	it('reports missing bindings and complete resolution and alias paths', () => {
		const container = new Container();
		const a = createServiceToken<number>('a');
		const b = createServiceToken<number>('b');

		expect(() => container.make(a)).toThrow(MissingBindingError);
		container.bind(a, (resolved) => resolved.make(b));
		container.bind(b, (resolved) => resolved.make(a));
		expect(() => container.make(a)).toThrow(CircularResolutionError);

		const c = createServiceToken<number>('c');

		container.alias(a, c);
		expect(() => container.alias(c, a)).toThrow(AliasCycleError);
	});

	it('keeps asynchronous resolution explicit and shares pending singleton work', async () => {
		const container = new Container();
		const asyncValue = createServiceToken<number>('async');

		let calls = 0;

		container.singletonAsync(asyncValue, async () => {
			calls += 1;

			await Promise.resolve();

			return 3;
		});
		expect(() => container.make(asyncValue)).toThrow(AsyncResolutionRequiredError);

		expect(await Promise.all([container.makeAsync(asyncValue), container.makeAsync(asyncValue)])).toEqual([3, 3]);

		expect(calls).toBe(1);
	});

	it('retries rejected async construction and detects cycles across awaits', async () => {
		const container = new Container();
		const retry = createServiceToken<number>('retry');

		let attempts = 0;

		container.singletonAsync(retry, async () => {
			attempts += 1;
			if (attempts === 1) {
				throw new Error('temporary');
			}

			return 2;
		});

		await expect(container.makeAsync(retry)).rejects.toThrow('temporary');

		expect(await container.makeAsync(retry)).toBe(2);

		const cycle = createServiceToken<number>('cycle');

		container.bindAsync(cycle, async (resolved) => await resolved.makeAsync(cycle));

		await expect(container.makeAsync(cycle)).rejects.toBeInstanceOf(CircularResolutionError);
	});

	it('runs provider lifecycle deterministically and isolates global application state', async () => {
		const service = createServiceToken<string>('service');
		const dependent = createServiceToken<string>('dependent');
		const events: string[] = [];

		const first = {
			provides: [service],
			register: (application: Application) => {
				events.push('register:first');
				application.instance(service, 'ready');
			},
			boot: () => {
				events.push('boot:first');
			},
		};

		const second = {
			provides: [dependent],
			dependsOn: [service],
			register: (application: Application) => {
				events.push('register:second');
				application.bind(dependent, (resolved) => resolved.make(service));
			},
			boot: () => {
				events.push('boot:second');
			},
		};

		const application = new Application();

		application.registerMany([second, first]).boot();
		expect(application.make(dependent)).toBe('ready');
		expect(events).toEqual(['register:first', 'register:second', 'boot:first', 'boot:second']);

		const deferred = createServiceToken<number>('deferred');

		await application.registerAsync({
			provides: [deferred],
			deferred: true,
			register: async (app) => {
				app.instance(deferred, 8);
			},
		});

		expect(await application.makeAsync(deferred)).toBe(8);

		setGlobalApplication(application);
		expect(getGlobalApplication()).toBe(application);
		clearGlobalApplication();
		expect(() => getGlobalApplication()).toThrow(MissingGlobalApplicationError);
	});

	it('rejects provider dependency cycles', () => {
		const a = createServiceToken<number>('provider-a');
		const b = createServiceToken<number>('provider-b');
		const application = new Application();

		expect(() =>
			application.registerMany([
				{ provides: [a], dependsOn: [b], register: () => undefined },
				{ provides: [b], dependsOn: [a], register: () => undefined },
			]),
		).toThrow(ProviderCycleError);
	});

	it('keeps tokens invariant by identity and preserves nullish cached values', () => {
		const container = new Container();
		const first = createServiceToken<number>('same-name');
		const second = createServiceToken<number>('same-name');

		const nil = createServiceToken<null>('nil', { guard: (value): value is null => value === null });

		const missing = createServiceToken<undefined>('missing', { guard: (value): value is undefined => value === undefined });

		let calls = 0;

		container.singleton(nil, () => {
			calls += 1;

			return null;
		});
		container.instance(missing, undefined);
		expect(container.make(nil)).toBeNull();
		expect(container.make(nil)).toBeNull();
		expect(calls).toBe(1);
		expect(container.make(missing)).toBeUndefined();
		expect(() => container.make(first)).toThrow(MissingBindingError);
		expect(() => container.make(second)).toThrow(MissingBindingError);
	});

	it('keeps parameters on precisely one factory and runs contextual values through lifecycle', () => {
		const container = new Container();
		const dependency = createServiceToken<string>('dependency');
		const dependencyAlias = createServiceToken<string>('dependency-alias');
		const parent = createServiceToken<string>('parent');
		const events: string[] = [];

		container.alias(dependency, dependencyAlias);
		container.bind(dependency, (resolved, parameters) => `${String(parameters.value)}:${String(resolved.parameters().value)}`);
		container.when(parent).needs(dependencyAlias).giveValue('contextual');
		container.resolving(dependency, (value) => events.push(`resolving:${value}`));
		container.afterResolving(dependency, (value) => events.push(`after:${value}`));
		container.bind(parent, (resolved, parameters) => `${String(parameters.value)}|${resolved.make(dependencyAlias)}`);

		expect(container.makeWith(parent, { value: 'parent' })).toBe('parent|contextual');
		expect(events).toEqual(['resolving:contextual', 'after:contextual']);
		expect(container.makeWith(dependency, { value: 'only-here' })).toBe('only-here:only-here');
	});

	it('skips missing tags, snapshots hooks, and never observes callbacks for a cache hit', () => {
		const container = new Container();
		const value = createServiceToken<number>('value');
		const absent = createServiceToken<number>('absent');
		const events: string[] = [];

		container.singleton(value, () => 1).tag([value, absent], 'all');
		container.beforeResolving(value, () => events.push('before'));
		container.resolving(value, () => events.push('resolving'));
		container.afterResolving(value, () => events.push('after'));
		expect(container.tagged('all')).toEqual([1]);
		expect(container.make(value)).toBe(1);
		expect(events).toEqual(['before', 'resolving', 'after', 'before']);
	});

	it('extends cached values exactly once and covers cached child scopes asynchronously', async () => {
		const container = new Container();
		const token = createServiceToken<number>('extended');
		const child = container.childScope();

		container.scoped(token, () => 1);
		expect(container.make(token)).toBe(1);
		expect(child.make(token)).toBe(1);
		container.extend(token, (value) => value + 1);
		expect(container.make(token)).toBe(2);
		expect(child.make(token)).toBe(2);

		await container.extendAsync(token, async (value) => value * 2);

		expect(container.make(token)).toBe(4);
		expect(child.make(token)).toBe(4);
		container.forgetScopedInstances();
		expect(container.make(token)).toBe(1);
	});

	it('isolates pending work, retries failures, and protects newer epochs from old completions', async () => {
		const container = new Container();
		const token = createServiceToken<string>('pending');

		let releaseOld: ((value: string) => void) | undefined;
		let releaseNew: ((value: string) => void) | undefined;

		container.singletonAsync(
			token,
			async () =>
				await new Promise<string>((resolve) => {
					releaseOld = resolve;
				}),
		);

		const old = container.makeAsync(token);

		await Promise.resolve();

		container.singletonAsync(
			token,
			async () =>
				await new Promise<string>((resolve) => {
					releaseNew = resolve;
				}),
		);

		const fresh = container.makeAsync(token);

		await Promise.resolve();

		releaseOld?.('old');
		releaseNew?.('new');

		expect(await old).toBe('old');

		expect(await fresh).toBe('new');

		expect(await container.makeAsync(token)).toBe('new');

		const scoped = createServiceToken<number>('scoped-pending');

		let calls = 0;

		container.scopedAsync(scoped, async () => {
			calls += 1;

			await Promise.resolve();

			return calls;
		});

		expect(await Promise.all([container.makeAsync(scoped), container.childScope().makeAsync(scoped)])).toEqual([2, 2]);

		expect(calls).toBe(2);
	});

	it('runs async observers for every pending caller and rejects sync access before known async hooks run', async () => {
		const container = new Container();
		const token = createServiceToken<number>('observers');
		const events: string[] = [];

		container.singletonAsync(token, async () => {
			await Promise.resolve();

			return 3;
		});
		container.resolvingAsync(token, async () => {
			events.push('resolving');
		});
		container.afterResolvingAsync(token, async () => {
			events.push('after');
		});

		await Promise.all([container.makeAsync(token), container.makeAsync(token)]);

		expect(events.filter((event) => event === 'resolving')).toHaveLength(2);
		expect(events.filter((event) => event === 'after')).toHaveLength(2);
		expect(container.make(token)).toBe(3);

		const uncached = new Container();
		const uncachedToken = createServiceToken<number>('uncached-async');

		uncached.singletonAsync(uncachedToken, async () => 9);
		expect(() => uncached.make(uncachedToken)).toThrow(AsyncResolutionRequiredError);

		const sync = new Container();
		const syncToken = createServiceToken<number>('sync');

		let called = false;

		sync
			.bind(syncToken, () => 1)
			.beforeResolvingAsync(syncToken, async () => {
				called = true;
			});
		expect(() => sync.make(syncToken)).toThrow(AsyncResolutionRequiredError);
		expect(called).toBe(false);
	});

	it('preflights async sync-resolution participants before factories, caches, or resolved mutation', async () => {
		const resolving = new Container();
		const resolvingToken = createServiceToken<number>('sync-resolving-async');

		let resolvingFactories = 0;

		resolving.singleton(resolvingToken, () => {
			resolvingFactories += 1;

			return 1;
		});
		resolving.resolvingAsync(resolvingToken, async () => undefined);
		expect(() => resolving.make(resolvingToken)).toThrow(AsyncResolutionRequiredError);
		expect(() => resolving.make(resolvingToken)).toThrow(AsyncResolutionRequiredError);
		expect(resolvingFactories).toBe(0);
		expect(resolving.resolved(resolvingToken)).toBe(false);

		const after = new Container();
		const afterToken = createServiceToken<number>('sync-after-async');

		let afterFactories = 0;

		after.singleton(afterToken, () => {
			afterFactories += 1;

			return 1;
		});
		after.afterResolvingAsync(afterToken, async () => undefined);
		expect(() => after.make(afterToken)).toThrow(AsyncResolutionRequiredError);
		expect(() => after.make(afterToken)).toThrow(AsyncResolutionRequiredError);
		expect(afterFactories).toBe(0);
		expect(after.resolved(afterToken)).toBe(false);

		const extender = new Container();
		const extenderToken = createServiceToken<number>('sync-extender-async');

		let extenderFactories = 0;
		let extenderHooks = 0;

		extender.bind(extenderToken, () => {
			extenderFactories += 1;

			return 1;
		});

		await extender.extendAsync(extenderToken, async (value) => {
			extenderHooks += 1;

			return value + 1;
		});

		expect(() => extender.make(extenderToken)).toThrow(AsyncResolutionRequiredError);
		expect(() => extender.make(extenderToken)).toThrow(AsyncResolutionRequiredError);
		expect(extenderFactories).toBe(0);
		expect(extenderHooks).toBe(0);
		expect(extender.resolved(extenderToken)).toBe(false);

		const asyncBinding = new Container();
		const asyncToken = createServiceToken<number>('async-binding-before');

		let beforeCalled = false;

		asyncBinding.bindAsync(asyncToken, async () => 1);
		asyncBinding.beforeResolving(asyncToken, () => {
			beforeCalled = true;
		});
		expect(() => asyncBinding.make(asyncToken)).toThrow(AsyncResolutionRequiredError);
		expect(beforeCalled).toBe(false);
	});

	it('applies per-caller callback snapshots while sharing pending singleton construction', async () => {
		const container = new Container();
		const token = createServiceToken<number>('per-caller');
		const events: string[] = [];

		let release: (() => void) | undefined;

		container.singletonAsync(
			token,
			async () =>
				await new Promise<number>((resolve) => {
					release = () => resolve(11);
				}),
		);

		const first = container.makeAsync(token);

		await Promise.resolve();

		await Promise.resolve();

		container.resolvingAsync(token, async () => {
			events.push('late');
		});

		const second = container.makeAsync(token);

		expect(release).toBeTypeOf('function');
		release?.();

		expect(await first).toBe(11);

		expect(await second).toBe(11);

		expect(events).toEqual(['late']);
	});

	it('forgets literal instance keys and extends cached values with guards and rebinding', async () => {
		const container = new Container();
		const target = createServiceToken<number>('forget-target');
		const alias = createServiceToken<number>('forget-alias');

		let builds = 0;

		container.singleton(target, () => {
			builds += 1;

			return builds;
		});
		expect(container.make(target)).toBe(1);
		container.alias(target, alias);
		container.forgetInstance(alias);
		expect(container.make(target)).toBe(1);
		expect(builds).toBe(1);

		const guarded = createServiceToken<number>('extend-guard', { guard: (value): value is number => typeof value === 'number' });

		const rebound: number[] = [];

		container.instance(guarded, 2);
		container.rebinding(guarded, (value) => rebound.push(value));
		container.extend(guarded, (value) => value + 3);
		expect(container.make(guarded)).toBe(5);
		expect(rebound).toEqual([2, 5]);
		expect(() => container.extend(guarded, () => 'bad' as never)).toThrow(IncorrectResolvedTypeError);
		expect(container.make(guarded)).toBe(5);

		await expect(container.extendAsync(guarded, async () => 'async-bad' as never)).rejects.toBeInstanceOf(IncorrectResolvedTypeError);

		expect(container.make(guarded)).toBe(5);
	});

	it('keeps rebinding synchronous so bind/instance cannot partially apply around async callbacks', () => {
		const container = new Container();
		const token = createServiceToken<number>('sync-rebinding-only');
		const seen: number[] = [];

		container.singleton(token, () => 1);
		container.rebinding(token, (value) => seen.push(value));
		container.singleton(token, () => 2);
		expect(seen).toEqual([1, 2]);
		expect(container.make(token)).toBe(2);
		expect('rebindingAsync' in container).toBe(false);
		expect('refreshAsync' in container).toBe(false);
	});

	it('supports build, factories, wraps, refreshes, methods, and global reset helpers', async () => {
		const container = new Container();
		const token = createServiceToken<number>('helpers');

		container.singleton(token, () => 1);
		expect(container.build((_resolved, parameters) => Number(parameters.value))).toBeNaN();
		expect(container.factoryFunc(token)()).toBe(1);

		expect(await container.factoryFuncAsync(token)()).toBe(1);

		expect(container.wrap((_resolved, parameters) => String(parameters.value), { value: 'wrapped' })()).toBe('wrapped');

		const refreshed: number[] = [];

		container.refresh(token, (value) => refreshed.push(value));
		container.singleton(token, () => 2);
		expect(refreshed).toEqual([1, 2]);
		expect(() => container.callMethodBinding('missing', null)).toThrow(MissingMethodBindingError);
		container.bindMethodAsync('async', async () => 7);
		expect(() => container.callMethodBinding('async', null)).toThrow(AsyncResolutionRequiredError);

		expect(await container.callMethodBindingAsync('async', null)).toBe(7);

		setInstance(undefined);
		expect(getInstance()).toBe(getInstance());

		const app = new Application();

		app.instance(token, 9);
		setGlobalApplication(app);
		expect(make(token)).toBe(9);
		clearGlobalApplication();
	});

	it('uses Go Kahn provider ordering and retries deferred async registration and boot', async () => {
		const dependency = createServiceToken<string>('dependency');
		const dependent = createServiceToken<string>('dependent');
		const unrelated = createServiceToken<string>('unrelated');
		const events: string[] = [];
		const app = new Application();

		await app.registerManyAsync([
			{
				provides: [dependent],
				dependsOn: [dependency],
				registerAsync: async () => {
					events.push('dependent');
				},
			},
			{
				provides: [dependency],
				registerAsync: async () => {
					events.push('dependency');
				},
			},
			{
				provides: [unrelated],
				registerAsync: async () => {
					events.push('unrelated');
				},
			},
		]);

		expect(events).toEqual(['dependency', 'unrelated', 'dependent']);

		const deferred = createServiceToken<number>('deferred-retry');

		let attempts = 0;

		await app.registerAsync({
			provides: [deferred],
			deferred: true,
			registerAsync: async (application) => {
				attempts += 1;
				if (attempts === 1) {
					throw new Error('retry');
				}

				application.instance(deferred, 4);
			},
		});

		await expect(app.makeAsync(deferred)).rejects.toThrow('retry');

		expect(await app.getAsync(deferred)).toBe(4);

		await app.bootAsync();

		expect(app.booted()).toBe(true);
	});

	it('flushes deferred providers through application get and nested dependency resolution', () => {
		const service = createServiceToken<string>('nested-deferred-service');
		const consumer = createServiceToken<string>('nested-deferred-consumer');
		const app = new Application();

		app.register({
			provides: [service],
			deferred: true,
			register(application) {
				application.instance(service, 'ready');
			},
		});
		app.bind(consumer, (resolved) => resolved.get(service));
		expect(app.get(consumer)).toBe('ready');
	});

	it('preserves Go instance overlays, alias identity, and shared-instance queries', () => {
		const container = new Container();
		const service = createServiceToken<string>('overlay-service');
		const target = createServiceToken<string>('overlay-target');
		const alias = createServiceToken<string>('overlay-alias');

		container.singleton(service, () => 'factory');
		container.instance(service, 'instance');
		expect(container.isShared(service)).toBe(true);
		expect(container.make(service)).toBe('instance');
		container.forgetInstance(service);
		expect(container.make(service)).toBe('factory');

		container.singleton(target, () => 'target');
		container.alias(target, alias);
		container.instance(alias, 'alias-instance');
		expect(container.isAlias(alias)).toBe(false);
		expect(container.make(alias)).toBe('alias-instance');
		expect(container.make(target)).toBe('target');
	});

	it('installs contextual bindings for every concrete and resolves typed configuration', () => {
		const container = new Container();
		const first = createServiceToken<string>('first-consumer');
		const second = createServiceToken<string>('second-consumer');
		const dependency = createServiceToken<number>('contextual-dependency');
		const config = createServiceToken<{ readonly timeout?: number }>('config');

		container.when(first, second).needs(dependency).give(7);
		container.bind(first, (resolved) => String(resolved.make(dependency)));
		container.bind(second, (resolved) => String(resolved.make(dependency)));
		expect(container.make(first)).toBe('7');
		expect(container.make(second)).toBe('7');

		container
			.when(first)
			.needs(dependency)
			.giveConfig(config, (value) => value.timeout, 30);
		expect(container.make(first)).toBe('30');
		container.instance(config, { timeout: 45 });
		expect(container.make(first)).toBe('45');
	});

	it('supports conditional async registrations without replacing existing bindings', async () => {
		const container = new Container();
		const bind = createServiceToken<number>('bind-async-if');
		const singleton = createServiceToken<number>('singleton-async-if');
		const scoped = createServiceToken<number>('scoped-async-if');

		container.bind(bind, () => 1).bindAsyncIf(bind, async () => 2);
		container.singletonAsyncIf(singleton, async () => 3);
		container.scopedAsyncIf(scoped, async () => 4);

		expect(await container.makeAsync(bind)).toBe(1);

		expect(await container.makeAsync(singleton)).toBe(3);

		expect(await container.makeAsync(scoped)).toBe(4);
	});

	it('awaits concurrent provider registration, deferred registration, boot, and late boot work', async () => {
		const registered = createServiceToken<string>('registered-concurrently');
		const deferred = createServiceToken<string>('deferred-concurrently');
		const events: string[] = [];
		const app = new Application();

		let releaseRegister: (() => void) | undefined;
		let registerCalls = 0;

		const provider: AsyncServiceProvider = {
			provides: [registered],
			registerAsync: async (application) => {
				registerCalls += 1;

				await new Promise<void>((resolve) => {
					releaseRegister = resolve;
				});

				application.instance(registered, 'registered');
				events.push('registered');
			},
		};

		const firstRegistration = app.registerAsync(provider);

		await Promise.resolve();

		await Promise.resolve();

		const secondRegistration = app.registerAsync(provider);

		expect(events).toEqual([]);
		expect(releaseRegister).toBeTypeOf('function');
		releaseRegister?.();

		await Promise.all([firstRegistration, secondRegistration]);

		expect(registerCalls).toBe(1);
		expect(events).toEqual(['registered']);

		let releaseDeferred: (() => void) | undefined;
		let deferredCalls = 0;

		await app.registerAsync({
			provides: [deferred],
			deferred: true,
			registerAsync: async (application) => {
				deferredCalls += 1;

				await new Promise<void>((resolve) => {
					releaseDeferred = resolve;
				});

				application.instance(deferred, 'deferred');
			},
		});

		const firstDeferred = app.makeAsync(deferred);

		await Promise.resolve();

		await Promise.resolve();

		const secondDeferred = app.makeAsync(deferred);

		expect(releaseDeferred).toBeTypeOf('function');
		releaseDeferred?.();

		expect(await Promise.all([firstDeferred, secondDeferred])).toEqual(['deferred', 'deferred']);

		expect(deferredCalls).toBe(1);

		let releaseBoot: (() => void) | undefined;
		let bootCalls = 0;

		await app.registerAsync({
			registerAsync: async () => undefined,
			bootAsync: async () => {
				bootCalls += 1;

				await new Promise<void>((resolve) => {
					releaseBoot = resolve;
				});
			},
		});

		expect(app.booted()).toBe(false);
		expect(app.providerStates().map((record) => record.registered)).toEqual([true, true, true]);

		const firstBoot = app.bootAsync();

		for (let tick = 0; tick < 4; tick += 1) {
			await Promise.resolve();
		}

		const secondBoot = app.bootAsync();

		expect(releaseBoot).toBeTypeOf('function');
		releaseBoot?.();

		await Promise.all([firstBoot, secondBoot]);

		expect(bootCalls).toBe(1);

		let releaseLateBoot: (() => void) | undefined;
		let lateBootCalls = 0;

		const late: AsyncServiceProvider = {
			registerAsync: async () => undefined,
			bootAsync: async () => {
				lateBootCalls += 1;

				await new Promise<void>((resolve) => {
					releaseLateBoot = resolve;
				});
			},
		};

		const firstLate = app.registerAsync(late);

		for (let tick = 0; tick < 4; tick += 1) {
			await Promise.resolve();
		}

		const secondLate = app.registerAsync(late);

		expect(releaseLateBoot).toBeTypeOf('function');
		releaseLateBoot?.();

		await Promise.all([firstLate, secondLate]);

		expect(lateBootCalls).toBe(1);
	});

	it('resets rejected provider lifecycle work and keeps invalid async provider state retryable', async () => {
		const app = new Application();

		let attempts = 0;

		const retrying: AsyncServiceProvider = {
			registerAsync: async () => {
				attempts += 1;
				if (attempts === 1) {
					throw new Error('registration failed');
				}
			},
		};

		await expect(app.registerAsync(retrying)).rejects.toThrow('registration failed');

		await expect(app.registerAsync(retrying)).resolves.toBe(app);

		// SAFETY: runtime validation must defend the public API against values
		// received from untyped JavaScript callers.
		const invalid = {} as unknown as AnyServiceProvider;

		await expect(app.registerAsync(invalid)).rejects.toThrow('must declare register or registerAsync');

		expect(app.providerStates().at(-1)?.registered).toBe(false);

		await expect(app.bootAsync()).resolves.toBe(app);

		const bootRetry = new Application();

		let bootAttempts = 0;

		await bootRetry.registerAsync({
			registerAsync: async () => undefined,
			bootAsync: async () => {
				bootAttempts += 1;
				if (bootAttempts === 1) {
					throw new Error('boot failed');
				}
			},
		});

		await expect(bootRetry.bootAsync()).rejects.toThrow('boot failed');

		await expect(bootRetry.bootAsync()).resolves.toBe(bootRetry);
	});

	it('retries late async provider boot through registerAsync and bootAsync with boot visibility', async () => {
		const lateBoot = new Application();
		const seen: boolean[] = [];

		await lateBoot.registerAsync({
			registerAsync: async () => undefined,
			bootAsync: async (application) => {
				seen.push(application.booted());
			},
		});

		await lateBoot.bootAsync();

		expect(seen).toEqual([true]);
		expect(lateBoot.booted()).toBe(true);

		let attempts = 0;

		const failing: AsyncServiceProvider = {
			registerAsync: async () => undefined,
			bootAsync: async () => {
				attempts += 1;
				if (attempts === 1) {
					throw new Error('late boot failed');
				}
			},
		};

		await expect(lateBoot.registerAsync(failing)).rejects.toThrow('late boot failed');

		expect(lateBoot.providerStates().at(-1)).toMatchObject({ registered: true, booted: false });
		expect(lateBoot.booted()).toBe(true);

		await expect(lateBoot.registerAsync(failing)).resolves.toBe(lateBoot);

		expect(lateBoot.providerStates().at(-1)?.booted).toBe(true);

		const retryBoot = new Application();

		let bootAttempts = 0;

		await retryBoot.bootAsync();

		await retryBoot
			.registerAsync({
				registerAsync: async () => undefined,
				bootAsync: async () => {
					bootAttempts += 1;
					if (bootAttempts === 1) {
						throw new Error('bootAsync retry');
					}
				},
			})
			.catch(() => undefined);

		expect(retryBoot.providerStates().at(-1)).toMatchObject({ registered: true, booted: false });

		await expect(retryBoot.bootAsync()).resolves.toBe(retryBoot);

		expect(retryBoot.providerStates().at(-1)?.booted).toBe(true);
		expect(bootAttempts).toBe(2);

		const pending = new Application();

		let releaseBoot: (() => void) | undefined;

		await pending.registerAsync({
			registerAsync: async () => undefined,
			bootAsync: async () => {
				await new Promise<void>((resolve) => {
					releaseBoot = resolve;
				});
			},
		});

		const pendingBoot = pending.bootAsync();

		for (let tick = 0; tick < 4; tick += 1) {
			await Promise.resolve();
		}

		expect(() => pending.boot()).toThrow(AsyncResolutionRequiredError);
		expect(pending.booted()).toBe(true);
		releaseBoot?.();

		await pendingBoot;

		const syncVisibility = new Application();

		let syncBooted = false;

		syncVisibility.register({
			register() {},
			boot(application) {
				syncBooted = application.booted();
			},
		});
		syncVisibility.boot();
		expect(syncBooted).toBe(true);

		const singleflight = new Application();

		let releaseRetry: (() => void) | undefined;
		let retryCalls = 0;

		await singleflight.registerAsync({
			registerAsync: async () => undefined,
			bootAsync: async () => {
				retryCalls += 1;

				await new Promise<void>((resolve) => {
					releaseRetry = resolve;
				});
			},
		});

		const firstRetry = singleflight.bootAsync();

		for (let tick = 0; tick < 4; tick += 1) {
			await Promise.resolve();
		}

		const secondRetry = singleflight.bootAsync();

		releaseRetry?.();

		await Promise.all([firstRetry, secondRetry]);

		expect(retryCalls).toBe(1);
	});

	it('snapshots resolution callbacks before factories and retries async callback failures without caching', async () => {
		const container = new Container();
		const snapshot = createServiceToken<number>('callback-snapshot');
		const events: string[] = [];

		container.bind(snapshot, (resolved) => {
			resolved.resolving(snapshot, () => events.push('late-resolving'));
			resolved.afterResolving(snapshot, () => events.push('late-after'));

			return 1;
		});
		expect(container.make(snapshot)).toBe(1);
		expect(events).toEqual([]);
		expect(container.make(snapshot)).toBe(1);
		expect(events).toEqual(['late-resolving', 'late-after']);

		const retry = createServiceToken<number>('async-callback-retry');

		let constructions = 0;
		let observerAttempts = 0;

		container.singletonAsync(retry, async () => ++constructions);
		container.resolvingAsync(retry, async () => {
			observerAttempts += 1;
			if (observerAttempts === 1) {
				throw new Error('observer failed');
			}
		});

		await expect(container.makeAsync(retry)).rejects.toThrow('observer failed');

		expect(await container.makeAsync(retry)).toBe(2);

		expect(constructions).toBe(2);

		const follower = createServiceToken<number>('async-follower-retry');

		let releaseFollower: (() => void) | undefined;
		let followerConstructions = 0;
		let followerObservers = 0;

		container.singletonAsync(follower, async () => {
			followerConstructions += 1;

			if (followerConstructions === 1) {
				await new Promise<void>((resolve) => {
					releaseFollower = resolve;
				});
			}

			return followerConstructions;
		});
		container.resolvingAsync(follower, async () => {
			followerObservers += 1;
			if (followerObservers === 2) {
				throw new Error('follower observer failed');
			}
		});

		const firstFollower = container.makeAsync(follower);
		const secondFollower = container.makeAsync(follower);

		await Promise.resolve();

		releaseFollower?.();

		await expect(firstFollower).resolves.toBe(1);

		await expect(secondFollower).rejects.toThrow('follower observer failed');

		expect(await container.makeAsync(follower)).toBe(2);
	});

	it('coordinates async graph cycles, sibling contexts, stale pending completions, and application child scopes', async () => {
		const container = new Container();
		const a = createServiceToken<number>('indirect-a');
		const b = createServiceToken<number>('indirect-b');
		const c = createServiceToken<number>('indirect-c');

		container.bindAsync(a, async (resolved) => await resolved.makeAsync(b));
		container.bindAsync(b, async (resolved) => await resolved.makeAsync(c));
		container.bindAsync(c, async (resolved) => await resolved.makeAsync(a));

		await expect(container.makeAsync(a)).rejects.toBeInstanceOf(CircularResolutionError);

		const left = createServiceToken<string>('parallel-left');
		const right = createServiceToken<string>('parallel-right');
		const parent = createServiceToken<string>('parallel-parent');

		container.bindAsync(left, async (resolved) => resolved.currentlyResolving()?.name ?? 'missing');
		container.bindAsync(right, async (resolved) => resolved.currentlyResolving()?.name ?? 'missing');
		container.bindAsync(parent, async (resolved) => (await Promise.all([resolved.makeAsync(left), resolved.makeAsync(right)])).join(','));

		expect(await container.makeAsync(parent)).toBe('parallel-left,parallel-right');

		const stale = createServiceToken<string>('stale-pending');

		let releaseStale: ((value: string) => void) | undefined;
		let staleCalls = 0;

		container.singletonAsync(stale, async () => {
			staleCalls += 1;

			if (staleCalls === 1) {
				return await new Promise<string>((resolve) => {
					releaseStale = resolve;
				});
			}

			return 'fresh';
		});

		const old = container.makeAsync(stale);

		await Promise.resolve();

		container.forgetInstance(stale);
		releaseStale?.('old');

		expect(await old).toBe('old');

		expect(await container.makeAsync(stale)).toBe('fresh');

		const flushed = new Container();
		const flushToken = createServiceToken<string>('flushed-pending');

		let releaseFlush: ((value: string) => void) | undefined;

		flushed.singletonAsync(
			flushToken,
			async () =>
				await new Promise<string>((resolve) => {
					releaseFlush = resolve;
				}),
		);

		const pendingFlush = flushed.makeAsync(flushToken);

		await Promise.resolve();

		flushed.flush();
		releaseFlush?.('stale');

		expect(await pendingFlush).toBe('stale');

		await expect(flushed.makeAsync(flushToken)).rejects.toBeInstanceOf(MissingBindingError);

		const app = new Application();
		const scoped = createServiceToken<object>('application-child-scope');

		app.scoped(scoped, () => ({}));

		const child = app.childScope();

		expect(child).toBeInstanceOf(Application);
		expect(child.make(scoped)).not.toBe(app.make(scoped));
	});
});

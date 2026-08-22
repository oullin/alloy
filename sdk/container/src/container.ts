import { serviceTokenGuard, type AnyServiceToken, type ServiceToken } from '#container/token';
import {
	AliasCycleError,
	AsyncResolutionRequiredError,
	CircularResolutionError,
	IncorrectResolvedTypeError,
	MissingBindingError,
	MissingMethodBindingError,
	SelfAliasError,
} from '#container/errors';

import type {
	AsyncExtender,
	AsyncFactory,
	AsyncMethodCallable,
	AsyncResolvingCallback,
	BindingOptions,
	Extender,
	Factory,
	Lifetime,
	MethodCallable,
	ResolutionParameters,
	ResolvingCallback,
} from '#container/types';

type Binding = { readonly factory: Factory<unknown> | AsyncFactory<unknown>; readonly lifetime: Lifetime; readonly asynchronous: boolean; readonly epoch: number };

type ResolutionCallbacks = { readonly resolving: readonly Callback[]; readonly afterResolving: readonly Callback[] };

type PendingValue = {
	readonly epoch: number;
	readonly promise: Promise<unknown>;
	participants: number;
	failed: boolean;
};

type SyncCallback = (value: unknown, container: Container) => void;

type AsyncCallback = (value: unknown, container: Container) => Promise<void>;

type Callback = { readonly callback: SyncCallback | AsyncCallback; readonly asynchronous: boolean };

type BeforeCallback = { readonly callback: (token: AnyServiceToken, parameters: ResolutionParameters, container: Container) => void | Promise<void>; readonly asynchronous: boolean };

type ContextualImplementation =
	| { readonly kind: 'value'; readonly value: unknown }
	| { readonly kind: 'factory'; readonly factory: Factory<unknown> }
	| { readonly kind: 'asyncFactory'; readonly factory: AsyncFactory<unknown> };

type Scope = { readonly instances: Map<AnyServiceToken, unknown>; readonly pending: Map<AnyServiceToken, PendingValue> };

type ResolutionContext = { readonly path: readonly AnyServiceToken[]; readonly parameters: ResolutionParameters };

type Core = {
	readonly bindings: Map<AnyServiceToken, Binding>;
	readonly aliases: Map<AnyServiceToken, AnyServiceToken>;
	readonly singletonInstances: Map<AnyServiceToken, unknown>;
	readonly singletonPending: Map<AnyServiceToken, PendingValue>;
	readonly epochs: Map<AnyServiceToken, number>;
	readonly resolved: Set<AnyServiceToken>;
	readonly contextual: Map<AnyServiceToken, Map<AnyServiceToken, ContextualImplementation>>;
	readonly tags: Map<string, AnyServiceToken[]>;
	readonly extenders: Map<AnyServiceToken, Callback[]>;
	readonly before: Map<AnyServiceToken | undefined, BeforeCallback[]>;
	readonly resolving: Map<AnyServiceToken | undefined, Callback[]>;
	readonly afterResolving: Map<AnyServiceToken | undefined, Callback[]>;
	readonly rebinding: Map<AnyServiceToken, Callback[]>;
	readonly methods: Map<string, { readonly callback: MethodCallable<unknown> | AsyncMethodCallable<unknown>; readonly asynchronous: boolean }>;
	readonly scopes: Set<Scope>;
};

const emptyParameters: ResolutionParameters = Object.freeze({});

const newCore = (): Core => ({
	bindings: new Map(),
	aliases: new Map(),
	singletonInstances: new Map(),
	singletonPending: new Map(),
	epochs: new Map(),
	resolved: new Set(),
	contextual: new Map(),
	tags: new Map(),
	extenders: new Map(),
	before: new Map(),
	resolving: new Map(),
	afterResolving: new Map(),
	rebinding: new Map(),
	methods: new Map(),
	scopes: new Set(),
});

const isThenable = (value: unknown): value is Promise<unknown> => typeof value === 'object' && value !== null && 'then' in value && typeof (value as { then?: unknown }).then === 'function';

const append = <T>(map: Map<AnyServiceToken | undefined, T[]>, token: AnyServiceToken | undefined, value: T): void => {
	map.set(token, [...(map.get(token) ?? []), value]);
};

/** An identity-token container with explicit synchronous and asynchronous APIs. */
export class Container {
	protected readonly core: Core;
	protected readonly scope: Scope;
	protected readonly context: ResolutionContext;

	/** Create an empty container. */
	constructor(core: Core = newCore(), scope: Scope = { instances: new Map(), pending: new Map() }, context: ResolutionContext = { path: [], parameters: emptyParameters }) {
		this.core = core;
		this.scope = scope;
		this.context = context;
		core.scopes.add(scope);
	}

	/** Create a TypeScript-only child scope with isolated scoped values. */
	childScope(): Container {
		return new Container(this.core);
	}
	/** Register a synchronous factory. */

	bind<T>(token: ServiceToken<T>, factory: Factory<T>, options: BindingOptions = {}): this {
		this.install(token, factory, false, options.lifetime ?? 'transient');

		return this;
	}
	/** Register an asynchronous factory. */

	bindAsync<T>(token: ServiceToken<T>, factory: AsyncFactory<T>, options: BindingOptions = {}): this {
		this.install(token, factory, true, options.lifetime ?? 'transient');

		return this;
	}
	/** Bind only when no value, binding, or alias exists. */
	bindIf<T>(token: ServiceToken<T>, factory: Factory<T>, options: BindingOptions = {}): this {
		if (!this.bound(token)) {
			this.bind(token, factory, options);
		}

		return this;
	}
	/** Register an asynchronous factory only when no value, binding, or alias exists. */
	bindAsyncIf<T>(token: ServiceToken<T>, factory: AsyncFactory<T>, options: BindingOptions = {}): this {
		if (!this.bound(token)) {
			this.bindAsync(token, factory, options);
		}

		return this;
	}
	/** Register a singleton factory. */
	singleton<T>(token: ServiceToken<T>, factory: Factory<T>): this {
		return this.bind(token, factory, { lifetime: 'singleton' });
	}
	/** Register a singleton factory only when unbound. */
	singletonIf<T>(token: ServiceToken<T>, factory: Factory<T>): this {
		if (!this.bound(token)) {
			this.singleton(token, factory);
		}

		return this;
	}
	/** Register an asynchronous singleton factory. */
	singletonAsync<T>(token: ServiceToken<T>, factory: AsyncFactory<T>): this {
		return this.bindAsync(token, factory, { lifetime: 'singleton' });
	}
	/** Register an asynchronous singleton factory only when unbound. */
	singletonAsyncIf<T>(token: ServiceToken<T>, factory: AsyncFactory<T>): this {
		if (!this.bound(token)) {
			this.singletonAsync(token, factory);
		}

		return this;
	}
	/** Register a scope-local factory. */
	scoped<T>(token: ServiceToken<T>, factory: Factory<T>): this {
		return this.bind(token, factory, { lifetime: 'scoped' });
	}
	/** Register a scope-local factory only when unbound. */
	scopedIf<T>(token: ServiceToken<T>, factory: Factory<T>): this {
		if (!this.bound(token)) {
			this.scoped(token, factory);
		}

		return this;
	}
	/** Register an asynchronous scope-local factory. */
	scopedAsync<T>(token: ServiceToken<T>, factory: AsyncFactory<T>): this {
		return this.bindAsync(token, factory, { lifetime: 'scoped' });
	}
	/** Register an asynchronous scope-local factory only when unbound. */
	scopedAsyncIf<T>(token: ServiceToken<T>, factory: AsyncFactory<T>): this {
		if (!this.bound(token)) {
			this.scopedAsync(token, factory);
		}

		return this;
	}
	/** Register an existing value, including null and undefined. */

	instance<T>(token: ServiceToken<T>, value: T): T {
		const supplied = token as AnyServiceToken;

		this.validate(supplied, value);

		const wasBound = this.bound(supplied);

		if (wasBound) {
			this.assertRebindingSync(supplied);
		}

		// Go keeps an existing factory beneath an explicit instance so
		// forgetInstance reveals it again. An alias is removed only at the
		// supplied name; its former target must remain unchanged.
		this.invalidate(supplied);
		this.core.aliases.delete(supplied);
		this.core.singletonInstances.set(supplied, value);
		this.core.resolved.add(supplied);
		if (wasBound) {
			this.notifyRebindingSync(supplied);
		}

		return value;
	}

	/** Create an alias that resolves by object identity. */
	alias<T>(abstract: ServiceToken<T>, name: ServiceToken<T>): this {
		if (abstract === name) {
			throw new SelfAliasError(name);
		}

		const path: AnyServiceToken[] = [name];

		let current: AnyServiceToken = abstract;

		while (true) {
			path.push(current);
			if (current === name) {
				throw new AliasCycleError(path);
			}

			const next = this.core.aliases.get(current);

			if (next === undefined) {
				break;
			}

			current = next;
		}

		this.core.aliases.set(name, abstract);

		return this;
	}
	/** Resolve an alias to its target. */
	getAlias<T>(token: ServiceToken<T>): ServiceToken<T> {
		return this.canonical(token) as ServiceToken<T>;
	}
	/** Return whether a token is directly an alias. */
	isAlias(token: AnyServiceToken): boolean {
		return this.core.aliases.has(token);
	}

	/** Resolve with an empty parameter map. */
	make<T>(token: ServiceToken<T>): T {
		return this.resolveSync(token, emptyParameters) as T;
	}
	/** Resolve the requested factory with supplied parameters. */
	makeWith<T>(token: ServiceToken<T>, parameters: ResolutionParameters): T {
		return this.resolveSync(token, parameters) as T;
	}
	/** Resolve an async-capable graph with an empty parameter map. */
	async makeAsync<T>(token: ServiceToken<T>): Promise<T> {
		return (await this.resolveAsync(token, emptyParameters)) as T;
	}
	/** Resolve the requested async-capable factory with supplied parameters. */
	async makeWithAsync<T>(token: ServiceToken<T>, parameters: ResolutionParameters): Promise<T> {
		return (await this.resolveAsync(token, parameters)) as T;
	}
	/** Resolve a factory directly without registering it. */
	build<T>(factory: Factory<T>, parameters: ResolutionParameters = emptyParameters): T {
		return this.call(factory, parameters);
	}
	/** Resolve an async-capable factory directly without registering it. */
	async buildAsync<T>(factory: Factory<T> | AsyncFactory<T>, parameters: ResolutionParameters = emptyParameters): Promise<T> {
		return await this.callAsync(factory, parameters);
	}
	/** Return a zero-argument resolver. */
	factoryFunc<T>(token: ServiceToken<T>): () => T {
		return () => this.make(token);
	}
	/** Return a zero-argument async resolver. */
	factoryFuncAsync<T>(token: ServiceToken<T>): () => Promise<T> {
		return async () => await this.makeAsync(token);
	}
	/** Resolve only an already-bound token. */
	get<T>(token: ServiceToken<T>): T {
		if (!this.bound(token)) {
			throw new MissingBindingError(token);
		}

		return this.make(token);
	}
	/** Resolve only an already-bound token asynchronously. */
	async getAsync<T>(token: ServiceToken<T>): Promise<T> {
		if (!this.bound(token)) {
			throw new MissingBindingError(token);
		}

		return await this.makeAsync(token);
	}
	/** Return the parameters visible to this active factory. */
	parameters(): ResolutionParameters {
		return this.context.parameters;
	}

	/** Begin a contextual binding declaration. */
	when(...concretes: readonly AnyServiceToken[]): ContextualBindingBuilder {
		return new ContextualBindingBuilder(
			this,
			concretes.map((concrete) => this.canonical(concrete)),
		);
	}
	/** Assign tokens to tag names. */
	tag(tokens: readonly AnyServiceToken[], ...tags: readonly string[]): this {
		for (const tag of tags) {
			this.core.tags.set(tag, [...(this.core.tags.get(tag) ?? []), ...tokens]);
		}

		return this;
	}
	/** Resolve a tag in registration order, skipping unresolvable entries. */

	tagged(tag: string): unknown[] {
		const values: unknown[] = [];

		for (const token of this.core.tags.get(tag) ?? []) {
			try {
				values.push(this.make(token as ServiceToken<unknown>));
			} catch (error) {
				if (!(error instanceof MissingBindingError)) {
					throw error;
				}
			}
		}

		return values;
	}
	/** Resolve an async-capable tag in registration order, skipping unresolvable entries. */

	async taggedAsync(tag: string): Promise<unknown[]> {
		const values: unknown[] = [];

		for (const token of this.core.tags.get(tag) ?? []) {
			try {
				values.push(await this.makeAsync(token as ServiceToken<unknown>));
			} catch (error) {
				if (!(error instanceof MissingBindingError)) {
					throw error;
				}
			}
		}

		return values;
	}

	/** Add a synchronous extender. Cached values receive only this new extender. */
	extend<T>(token: ServiceToken<T>, extender: Extender<T>): this {
		const canonical = this.canonical(token);
		const cached = this.cachedEntries(canonical);

		if (cached.length > 0) {
			const updated = cached.map((entry) => {
				const extended = this.applyOneSync(extender as Extender<unknown>, entry.value, this);

				this.validate(canonical, extended);

				return { entry, extended };
			});

			for (const { entry, extended } of updated) {
				entry.set(extended);
			}

			this.notifyRebindingSync(canonical);

			return this;
		}

		append(this.core.extenders, canonical, { callback: extender as SyncCallback, asynchronous: false });
		if (this.core.resolved.has(canonical)) {
			this.notifyRebindingSync(canonical);
		}

		return this;
	}
	/** Add an async extender. Cached values in every existing scope receive it now. */
	async extendAsync<T>(token: ServiceToken<T>, extender: AsyncExtender<T>): Promise<this> {
		const canonical = this.canonical(token);
		const cached = this.cachedEntries(canonical);

		if (cached.length > 0) {
			const updated: { readonly entry: (typeof cached)[number]; readonly extended: unknown }[] = [];

			for (const entry of cached) {
				const extended = await (extender as AsyncExtender<unknown>)(entry.value, this);

				this.validate(canonical, extended);
				updated.push({ entry, extended });
			}

			for (const { entry, extended } of updated) {
				entry.set(extended);
			}

			this.notifyRebindingSync(canonical);

			return this;
		}

		append(this.core.extenders, canonical, { callback: extender as AsyncCallback, asynchronous: true });
		if (this.core.resolved.has(canonical)) {
			this.notifyRebindingSync(canonical);
		}

		return this;
	}
	/** Remove all future extenders for a token. */
	forgetExtenders(token: AnyServiceToken): void {
		this.core.extenders.delete(this.canonical(token));
	}

	/** Register a synchronous pre-resolution callback. */

	beforeResolving<T>(token: ServiceToken<T> | undefined, callback: (token: ServiceToken<T>, parameters: ResolutionParameters, container: Container) => void): this {
		append(this.core.before, token, { callback: callback as BeforeCallback['callback'], asynchronous: false });

		return this;
	}
	/** Register an asynchronous pre-resolution callback. */

	beforeResolvingAsync<T>(token: ServiceToken<T> | undefined, callback: (token: ServiceToken<T>, parameters: ResolutionParameters, container: Container) => Promise<void>): this {
		append(this.core.before, token, { callback: callback as BeforeCallback['callback'], asynchronous: true });

		return this;
	}
	/** Register a synchronous post-construction callback. */

	resolving<T>(token: ServiceToken<T> | undefined, callback: ResolvingCallback<T>): this {
		append(this.core.resolving, token, { callback: callback as SyncCallback, asynchronous: false });

		return this;
	}
	/** Register an asynchronous post-construction callback. */

	resolvingAsync<T>(token: ServiceToken<T> | undefined, callback: AsyncResolvingCallback<T>): this {
		append(this.core.resolving, token, { callback: callback as AsyncCallback, asynchronous: true });

		return this;
	}
	/** Register a synchronous final-resolution callback. */

	afterResolving<T>(token: ServiceToken<T> | undefined, callback: ResolvingCallback<T>): this {
		append(this.core.afterResolving, token, { callback: callback as SyncCallback, asynchronous: false });

		return this;
	}
	/** Register an asynchronous final-resolution callback. */

	afterResolvingAsync<T>(token: ServiceToken<T> | undefined, callback: AsyncResolvingCallback<T>): this {
		append(this.core.afterResolving, token, { callback: callback as AsyncCallback, asynchronous: true });

		return this;
	}
	/**
	 * Register a synchronous replacement callback and resolve immediately when already bound.
	 *
	 * Rebinding and refresh are synchronous Go-parity APIs. Asynchronous resolving
	 * callbacks, extenders, and methods remain supported TypeScript extensions; there is
	 * no async rebinding/refresh subscription because a later sync bind/instance would
	 * otherwise mutate state and then fail mid-notification.
	 */
	rebinding<T>(token: ServiceToken<T>, callback: ResolvingCallback<T>): T | undefined {
		const canonical = this.canonical(token);

		append(this.core.rebinding, canonical, { callback: callback as SyncCallback, asynchronous: false });
		if (!this.bound(canonical)) {
			return undefined;
		}

		const value = this.make(canonical as ServiceToken<T>);

		callback(value, this);

		return value;
	}
	/** Keep an external target refreshed whenever a token is rebound. */
	refresh<T>(token: ServiceToken<T>, setter: (value: T) => void): T | undefined {
		return this.rebinding(token, (value) => setter(value));
	}

	/** Bind a named synchronous method. */

	bindMethod<T>(method: string, callback: MethodCallable<T>): this {
		this.core.methods.set(method, { callback: callback as MethodCallable<unknown>, asynchronous: false });

		return this;
	}
	/** Bind a named asynchronous method. */

	bindMethodAsync<T>(method: string, callback: AsyncMethodCallable<T>): this {
		this.core.methods.set(method, { callback: callback as AsyncMethodCallable<unknown>, asynchronous: true });

		return this;
	}
	/** Invoke a method callback. */

	call<T>(callback: MethodCallable<T>, parameters: ResolutionParameters = emptyParameters): T {
		const result = callback(this.withContext(this.context.path, parameters), parameters);

		if (isThenable(result)) {
			throw new AsyncResolutionRequiredError('method callback');
		}

		return result;
	}
	/** Invoke an async-capable method callback. */
	async callAsync<T>(callback: MethodCallable<T> | AsyncMethodCallable<T>, parameters: ResolutionParameters = emptyParameters): Promise<T> {
		return await callback(this.withContext(this.context.path, parameters), parameters);
	}
	/** Return a deferred synchronous method invocation. */
	wrap<T>(callback: MethodCallable<T>, parameters: ResolutionParameters = emptyParameters): () => T {
		return () => this.call(callback, parameters);
	}
	/** Return a deferred async method invocation. */
	wrapAsync<T>(callback: MethodCallable<T> | AsyncMethodCallable<T>, parameters: ResolutionParameters = emptyParameters): () => Promise<T> {
		return async () => await this.callAsync(callback, parameters);
	}
	/** Invoke a named method binding. */

	callMethodBinding<T>(method: string, instance: unknown): T {
		const record = this.core.methods.get(method);

		if (record === undefined) {
			throw new MissingMethodBindingError(method);
		}

		if (record.asynchronous) {
			throw new AsyncResolutionRequiredError(`method binding "${method}"`);
		}

		return this.call(record.callback as MethodCallable<T>, { _instance: instance });
	}
	/** Invoke a named async-capable method binding. */

	async callMethodBindingAsync<T>(method: string, instance: unknown): Promise<T> {
		const record = this.core.methods.get(method);

		if (record === undefined) {
			throw new MissingMethodBindingError(method);
		}

		return await this.callAsync(record.callback as MethodCallable<T> | AsyncMethodCallable<T>, { _instance: instance });
	}
	/** Return whether a named method exists. */
	hasMethodBinding(method: string): boolean {
		return this.core.methods.has(method);
	}

	/** Return whether a token has a binding, instance, or alias. */

	bound(token: AnyServiceToken): boolean {
		const canonical = this.canonical(token);

		return this.core.bindings.has(canonical) || this.core.singletonInstances.has(canonical) || this.scope.instances.has(canonical) || this.core.aliases.has(token);
	}
	/** PSR-11 alias for bound. */
	has(token: AnyServiceToken): boolean {
		return this.bound(token);
	}
	/** Return whether a token has been resolved. */
	resolved(token: AnyServiceToken): boolean {
		return this.core.resolved.has(this.canonical(token));
	}
	/** Return whether the current binding is shared. */

	isShared(token: AnyServiceToken): boolean {
		const canonical = this.canonical(token);
		const binding = this.core.bindings.get(canonical);

		return this.core.singletonInstances.has(canonical) || (binding?.lifetime !== undefined && binding.lifetime !== 'transient');
	}
	/** Return the currently constructing token. */
	currentlyResolving(): AnyServiceToken | undefined {
		return this.context.path.at(-1);
	}
	/** Return binding metadata. */
	bindings(): ReadonlyMap<AnyServiceToken, Readonly<{ lifetime: Lifetime; asynchronous: boolean }>> {
		return new Map([...this.core.bindings].map(([token, binding]) => [token, { lifetime: binding.lifetime, asynchronous: binding.asynchronous }]));
	}
	/** Forget one cached value and pending construction at the supplied token key. */
	forgetInstance(token: AnyServiceToken): void {
		// Go deletes the literal instances map key without alias canonicalization.
		this.bump(token);
		this.core.singletonInstances.delete(token);
		this.core.singletonPending.delete(token);

		for (const scope of this.core.scopes) {
			scope.instances.delete(token);
			scope.pending.delete(token);
		}
	}
	/** Forget singleton and scoped cached values and pending construction. */
	forgetInstances(): void {
		for (const token of new Set([...this.core.singletonInstances.keys(), ...this.core.singletonPending.keys(), ...this.core.bindings.keys()])) {
			this.invalidate(token);
		}
	}
	/** Forget only scoped cached values and pending construction. */
	forgetScopedInstances(): void {
		for (const [token, binding] of this.core.bindings) {
			if (binding.lifetime === 'scoped') {
				this.invalidateScoped(token);
			}
		}
	}
	/** Reset bindings, aliases, hooks, methods, and cached state. Epochs intentionally survive. */
	flush(): void {
		for (const token of new Set([...this.core.bindings.keys(), ...this.core.singletonInstances.keys(), ...this.core.singletonPending.keys()])) {
			this.invalidate(token);
		}

		this.core.bindings.clear();
		this.core.aliases.clear();
		this.core.resolved.clear();
		this.core.contextual.clear();
		this.core.tags.clear();
		this.core.extenders.clear();
		this.core.before.clear();
		this.core.resolving.clear();
		this.core.afterResolving.clear();
		this.core.rebinding.clear();
		this.core.methods.clear();
	}

	protected resolveSync(token: AnyServiceToken, parameters: ResolutionParameters): unknown {
		const canonical = this.canonical(token);
		const path = this.context.path;

		if (path.includes(canonical)) {
			throw new CircularResolutionError([...path, canonical]);
		}

		const callbacks = this.snapshotResolutionCallbacks(canonical);
		const contextual = this.contextual(path.at(-1), canonical, token);
		const binding = this.core.bindings.get(canonical);
		const emptyParameters = Object.keys(parameters).length === 0;
		const cacheHit = contextual === undefined && emptyParameters && this.hasCached(canonical);

		if (cacheHit) {
			this.preflightBeforeSync(canonical);
			this.fireBeforeSync(canonical, parameters);

			return this.cached(canonical);
		}

		if (contextual?.kind === 'asyncFactory') {
			throw new AsyncResolutionRequiredError(canonical.name);
		}

		if (contextual === undefined && binding?.asynchronous === true) {
			throw new AsyncResolutionRequiredError(canonical.name);
		}

		this.preflightConstructionSync(canonical, callbacks);
		this.fireBeforeSync(canonical, parameters);

		if (contextual !== undefined) {
			return this.finishSync(canonical, this.constructContextualSync(canonical, contextual, parameters, path), parameters, path, callbacks);
		}

		if (binding === undefined) {
			throw new MissingBindingError(canonical);
		}

		return this.finishSync(canonical, this.constructSync(canonical, binding, parameters, path), parameters, path, callbacks);
	}

	protected async resolveAsync(token: AnyServiceToken, parameters: ResolutionParameters): Promise<unknown> {
		const canonical = this.canonical(token);
		const path = this.context.path;

		if (path.includes(canonical)) {
			throw new CircularResolutionError([...path, canonical]);
		}

		const callbacks = this.snapshotResolutionCallbacks(canonical);

		await this.fireBeforeAsync(canonical, parameters);

		const contextual = this.contextual(path.at(-1), canonical, token);

		if (contextual !== undefined) {
			return await this.finishAsync(canonical, await this.constructContextualAsync(canonical, contextual, parameters, path), parameters, path, callbacks);
		}

		if (Object.keys(parameters).length === 0 && this.hasCached(canonical)) {
			return this.cached(canonical);
		}

		const binding = this.core.bindings.get(canonical);

		if (binding === undefined) {
			throw new MissingBindingError(canonical);
		}

		if (binding.lifetime === 'transient' || Object.keys(parameters).length > 0) {
			return await this.finishAsync(canonical, await this.constructAsync(canonical, binding, parameters, path), parameters, path, callbacks);
		}

		const current = this.pending(canonical);
		const pending = current !== undefined && current.epoch === binding.epoch ? current : this.startPending(canonical, binding, parameters, path);

		pending.participants += 1;

		return await this.finishPendingAsync(canonical, binding, pending, parameters, path, callbacks);
	}

	private constructSync(token: AnyServiceToken, binding: Binding, parameters: ResolutionParameters, path: readonly AnyServiceToken[]): unknown {
		const child = this.withContext([...path, token], parameters);
		const extenders = [...(this.core.extenders.get(token) ?? [])];
		const value = (binding.factory as Factory<unknown>)(child, parameters);

		if (isThenable(value)) {
			throw new AsyncResolutionRequiredError(token.name);
		}

		return this.constructValueSync(token, value, child, extenders, binding, parameters);
	}

	private constructValueSync(
		token: AnyServiceToken,
		value: unknown,
		child: Container,
		extenders: readonly Callback[],
		binding: Binding | undefined,
		parameters: ResolutionParameters,
	): unknown {
		let extended = value;

		for (const extender of extenders) {
			if (extender.asynchronous) {
				throw new AsyncResolutionRequiredError(`extender for ${token.name}`);
			}

			extended = (extender.callback as SyncCallback)(extended, child);
		}

		this.validate(token, extended);
		if (binding !== undefined && binding.lifetime !== 'transient' && Object.keys(parameters).length === 0 && this.epoch(token) === binding.epoch) {
			this.cache(token, binding, extended);
		}

		return extended;
	}

	private async constructAsync(token: AnyServiceToken, binding: Binding, parameters: ResolutionParameters, path: readonly AnyServiceToken[]): Promise<unknown> {
		const child = this.withContext([...path, token], parameters);
		const extenders = [...(this.core.extenders.get(token) ?? [])];

		let extended = await binding.factory(child, parameters);

		for (const extender of extenders) {
			extended = extender.asynchronous ? await (extender.callback as AsyncCallback)(extended, child) : (extender.callback as SyncCallback)(extended, child);
		}

		this.validate(token, extended);

		return extended;
	}

	private constructContextualSync(token: AnyServiceToken, implementation: ContextualImplementation, parameters: ResolutionParameters, path: readonly AnyServiceToken[]): unknown {
		const child = this.withContext([...path, token], parameters);

		if (implementation.kind === 'asyncFactory') {
			throw new AsyncResolutionRequiredError(token.name);
		}

		const value = implementation.kind === 'value' ? implementation.value : implementation.factory(child, parameters);

		if (isThenable(value)) {
			throw new AsyncResolutionRequiredError(token.name);
		}

		return this.constructValueSync(token, value, child, [...(this.core.extenders.get(token) ?? [])], undefined, parameters);
	}

	private async constructContextualAsync(
		token: AnyServiceToken,
		implementation: ContextualImplementation,
		parameters: ResolutionParameters,
		path: readonly AnyServiceToken[],
	): Promise<unknown> {
		const child = this.withContext([...path, token], parameters);

		const value = implementation.kind === 'value' ? implementation.value : await implementation.factory(child, parameters);

		let extended = value;

		for (const extender of this.core.extenders.get(token) ?? []) {
			extended = extender.asynchronous ? await (extender.callback as AsyncCallback)(extended, child) : (extender.callback as SyncCallback)(extended, child);
		}

		this.validate(token, extended);

		return extended;
	}

	private finishSync(token: AnyServiceToken, value: unknown, _parameters: ResolutionParameters, path: readonly AnyServiceToken[], callbacks: ResolutionCallbacks): unknown {
		this.core.resolved.add(token);

		const child = this.withContext([...path, token], _parameters);

		this.fireCallbackSnapshotSync(callbacks.resolving, token, value, child);
		this.fireCallbackSnapshotSync(callbacks.afterResolving, token, value, child);

		return value;
	}

	private async finishAsync(token: AnyServiceToken, value: unknown, parameters: ResolutionParameters, path: readonly AnyServiceToken[], callbacks: ResolutionCallbacks): Promise<unknown> {
		const child = this.withContext([...path, token], parameters);

		await this.fireCallbackSnapshotAsync(callbacks.resolving, value, child);

		await this.fireCallbackSnapshotAsync(callbacks.afterResolving, value, child);

		this.core.resolved.add(token);

		return value;
	}

	private startPending(token: AnyServiceToken, binding: Binding, parameters: ResolutionParameters, path: readonly AnyServiceToken[]): PendingValue {
		const promise = this.constructAsync(token, binding, parameters, path);
		const pending: PendingValue = { epoch: binding.epoch, promise, participants: 0, failed: false };

		this.setPending(token, binding, pending);

		return pending;
	}

	private async finishPendingAsync(
		token: AnyServiceToken,
		binding: Binding,
		pending: PendingValue,
		parameters: ResolutionParameters,
		path: readonly AnyServiceToken[],
		callbacks: ResolutionCallbacks,
	): Promise<unknown> {
		try {
			const value = await pending.promise;

			return await this.finishAsync(token, value, parameters, path, callbacks);
		} catch (error) {
			pending.failed = true;
			throw error;
		} finally {
			pending.participants -= 1;

			if (pending.participants === 0) {
				if (!pending.failed && this.pending(token) === pending && this.epoch(token) === binding.epoch) {
					this.cache(token, binding, await pending.promise);
				}

				if (this.pending(token) === pending) {
					this.deletePending(token, binding);
				}
			}
		}
	}

	private install<T>(token: ServiceToken<T>, factory: Factory<T> | AsyncFactory<T>, asynchronous: boolean, lifetime: Lifetime): void {
		const canonical = token as AnyServiceToken;
		const wasResolved = this.core.resolved.has(canonical);

		if (wasResolved) {
			this.assertRebindingSync(canonical);
		}

		this.invalidate(canonical);
		this.core.aliases.delete(canonical);

		const epoch = this.epoch(canonical);

		this.core.bindings.set(canonical, { factory: factory as Factory<unknown> | AsyncFactory<unknown>, lifetime, asynchronous, epoch });
		if (wasResolved) {
			this.notifyRebindingSync(canonical);
		}
	}

	private canonical(token: AnyServiceToken): AnyServiceToken {
		const path: AnyServiceToken[] = [token];

		let current = token;

		while (this.core.aliases.has(current)) {
			const next = this.core.aliases.get(current);

			if (next === undefined) {
				break;
			}

			if (path.includes(next)) {
				throw new AliasCycleError([...path, next]);
			}

			path.push(next);
			current = next;
		}

		return current;
	}

	protected withContext(path: readonly AnyServiceToken[], parameters: ResolutionParameters): Container {
		return new Container(this.core, this.scope, { path: Object.freeze([...path]), parameters });
	}

	private contextual(requester: AnyServiceToken | undefined, canonical: AnyServiceToken, requested: AnyServiceToken): ContextualImplementation | undefined {
		if (requester === undefined) {
			return undefined;
		}

		const table = this.core.contextual.get(requester);

		if (table === undefined) {
			return undefined;
		}

		return table.has(canonical) ? table.get(canonical) : table.get(requested);
	}

	private epoch(token: AnyServiceToken): number {
		return this.core.epochs.get(token) ?? 0;
	}

	private bump(token: AnyServiceToken): number {
		const next = this.epoch(token) + 1;

		this.core.epochs.set(token, next);

		return next;
	}

	private hasCached(token: AnyServiceToken): boolean {
		return this.scope.instances.has(token) || this.core.singletonInstances.has(token);
	}

	private cached(token: AnyServiceToken): unknown {
		return this.scope.instances.has(token) ? this.scope.instances.get(token) : this.core.singletonInstances.get(token);
	}

	private pending(token: AnyServiceToken): PendingValue | undefined {
		return this.scope.pending.get(token) ?? this.core.singletonPending.get(token);
	}

	private setPending(token: AnyServiceToken, binding: Binding, pending: PendingValue): void {
		if (binding.lifetime === 'scoped') {
			this.scope.pending.set(token, pending);
		} else {
			this.core.singletonPending.set(token, pending);
		}
	}

	private deletePending(token: AnyServiceToken, binding: Binding): void {
		if (binding.lifetime === 'scoped') {
			this.scope.pending.delete(token);
		} else {
			this.core.singletonPending.delete(token);
		}
	}

	private cache(token: AnyServiceToken, binding: Binding, value: unknown): void {
		if (binding.lifetime === 'scoped') {
			this.scope.instances.set(token, value);
		} else {
			this.core.singletonInstances.set(token, value);
		}
	}

	private cachedEntries(token: AnyServiceToken): { readonly value: unknown; readonly set: (value: unknown) => void }[] {
		const entries: { value: unknown; set: (value: unknown) => void }[] = [];

		if (this.core.singletonInstances.has(token)) {
			entries.push({ value: this.core.singletonInstances.get(token), set: (value) => this.core.singletonInstances.set(token, value) });
		}
		for (const scope of this.core.scopes) {
			if (scope.instances.has(token)) {
				entries.push({ value: scope.instances.get(token), set: (value) => scope.instances.set(token, value) });
			}
		}

		return entries;
	}

	private invalidate(token: AnyServiceToken): void {
		this.bump(token);
		this.core.singletonInstances.delete(token);
		this.core.singletonPending.delete(token);

		for (const scope of this.core.scopes) {
			scope.instances.delete(token);
			scope.pending.delete(token);
		}
	}

	private invalidateScoped(token: AnyServiceToken): void {
		this.bump(token);

		for (const scope of this.core.scopes) {
			scope.instances.delete(token);
			scope.pending.delete(token);
		}
	}

	private validate(token: AnyServiceToken, value: unknown): void {
		const guard = serviceTokenGuard(token);

		if (guard !== undefined && !guard(value)) {
			throw new IncorrectResolvedTypeError(token, value);
		}
	}

	private applyOneSync(extender: Extender<unknown>, value: unknown, container: Container): unknown {
		const extended = extender(value, container);

		if (isThenable(extended)) {
			throw new AsyncResolutionRequiredError('extender');
		}

		return extended;
	}

	private callbacks(map: Map<AnyServiceToken | undefined, Callback[]>, token: AnyServiceToken): readonly Callback[] {
		return [...(map.get(undefined) ?? []), ...(map.get(token) ?? [])];
	}

	private snapshotResolutionCallbacks(token: AnyServiceToken): ResolutionCallbacks {
		return { resolving: this.callbacks(this.core.resolving, token), afterResolving: this.callbacks(this.core.afterResolving, token) };
	}

	private beforeCallbacks(token: AnyServiceToken): readonly BeforeCallback[] {
		return [...(this.core.before.get(undefined) ?? []), ...(this.core.before.get(token) ?? [])];
	}

	private preflightBeforeSync(token: AnyServiceToken): void {
		for (const entry of this.beforeCallbacks(token)) {
			if (entry.asynchronous) {
				throw new AsyncResolutionRequiredError(`before-resolving callback for ${token.name}`);
			}
		}
	}

	private preflightConstructionSync(token: AnyServiceToken, callbacks: ResolutionCallbacks): void {
		this.preflightBeforeSync(token);

		for (const entry of callbacks.resolving) {
			if (entry.asynchronous) {
				throw new AsyncResolutionRequiredError(`resolution callback for ${token.name}`);
			}
		}

		for (const entry of callbacks.afterResolving) {
			if (entry.asynchronous) {
				throw new AsyncResolutionRequiredError(`resolution callback for ${token.name}`);
			}
		}

		for (const extender of this.core.extenders.get(token) ?? []) {
			if (extender.asynchronous) {
				throw new AsyncResolutionRequiredError(`extender for ${token.name}`);
			}
		}
	}

	private fireBeforeSync(token: AnyServiceToken, parameters: ResolutionParameters): void {
		for (const entry of this.beforeCallbacks(token)) {
			if (entry.asynchronous) {
				throw new AsyncResolutionRequiredError(`before-resolving callback for ${token.name}`);
			}

			const result = entry.callback(token, parameters, this);

			if (isThenable(result)) {
				throw new AsyncResolutionRequiredError(`before-resolving callback for ${token.name}`);
			}
		}
	}

	private async fireBeforeAsync(token: AnyServiceToken, parameters: ResolutionParameters): Promise<void> {
		for (const entry of this.beforeCallbacks(token)) {
			await entry.callback(token, parameters, this);
		}
	}

	private fireCallbackSnapshotSync(callbacks: readonly Callback[], token: AnyServiceToken, value: unknown, container: Container): void {
		for (const entry of callbacks) {
			if (entry.asynchronous) {
				throw new AsyncResolutionRequiredError(`resolution callback for ${token.name}`);
			}

			const result = (entry.callback as SyncCallback)(value, container);

			if (isThenable(result)) {
				throw new AsyncResolutionRequiredError(`resolution callback for ${token.name}`);
			}
		}
	}

	private async fireCallbackSnapshotAsync(callbacks: readonly Callback[], value: unknown, container: Container): Promise<void> {
		for (const entry of callbacks) {
			await entry.callback(value, container);
		}
	}

	private assertRebindingSync(token: AnyServiceToken): void {
		for (const entry of this.core.rebinding.get(token) ?? []) {
			if (entry.asynchronous) {
				throw new AsyncResolutionRequiredError(`rebinding callback for ${token.name}`);
			}
		}
	}

	private notifyRebindingSync(token: AnyServiceToken): void {
		const callbacks = [...(this.core.rebinding.get(token) ?? [])];

		if (callbacks.length === 0) {
			return;
		}

		this.assertRebindingSync(token);

		const value = this.make(token as ServiceToken<unknown>);

		for (const entry of callbacks) {
			if (entry.asynchronous) {
				throw new AsyncResolutionRequiredError(`rebinding callback for ${token.name}`);
			}

			const result = (entry.callback as SyncCallback)(value, this);

			if (isThenable(result)) {
				throw new AsyncResolutionRequiredError(`rebinding callback for ${token.name}`);
			}
		}
	}
	/** @internal Store an explicit contextual implementation. */

	addContextualBinding(concrete: AnyServiceToken, dependency: AnyServiceToken, implementation: ContextualImplementation): void {
		const table = this.core.contextual.get(concrete) ?? new Map();

		table.set(dependency, implementation);
		this.core.contextual.set(concrete, table);
	}
}

/** Fluent builder for context-specific token implementations. */
export class ContextualBindingBuilder {
	constructor(
		private readonly container: Container,
		private readonly concretes: readonly AnyServiceToken[],
	) {}
	/** Select a dependency token. */
	needs<T>(token: ServiceToken<T>): ContextualBindingBuilderFor<T> {
		return new ContextualBindingBuilderFor<T>(this.container, this.concretes, token);
	}
}

/** A contextual binding builder after its dependency type has been selected. */
export class ContextualBindingBuilderFor<T> extends ContextualBindingBuilder {
	constructor(
		private readonly bindingContainer: Container,
		private readonly bindingConcretes: readonly AnyServiceToken[],
		private readonly dependency: ServiceToken<T>,
	) {
		super(bindingContainer, bindingConcretes);
	}
	/** Supply a static value, including a callable value. */
	giveValue(value: T): void {
		this.set({ kind: 'value', value });
	}
	/** Supply a synchronous contextual factory. */
	giveFactory(factory: Factory<T>): void {
		this.set({ kind: 'factory', factory: factory as Factory<unknown> });
	}
	/** Supply an asynchronous contextual factory. */
	giveAsyncFactory(factory: AsyncFactory<T>): void {
		this.set({ kind: 'asyncFactory', factory: factory as AsyncFactory<unknown> });
	}
	/** Backwards-compatible value spelling; callable values stay values. */
	give(value: T): void {
		this.giveValue(value);
	}
	/** Supply values resolved from a tag. */
	giveTagged(tag: string): void {
		this.set({ kind: 'factory', factory: (container) => container.tagged(tag) });
	}
	/**
	 * Supply a typed configuration value from an explicit configuration token.
	 *
	 * The getter keeps TypeScript's token identity boundary explicit. If the
	 * configuration token is unavailable or the getter returns undefined, the
	 * optional fallback is used; without a fallback the original resolution
	 * failure or undefined result is preserved.
	 */
	giveConfig<C>(config: ServiceToken<C>, getter: (value: C) => T | undefined, fallback?: T): void {
		this.set({
			kind: 'factory',
			factory: (container) => {
				try {
					const value = getter(container.make(config));

					return value === undefined ? fallback : value;
				} catch (error) {
					if (fallback !== undefined) {
						return fallback;
					}

					throw error;
				}
			},
		});
	}

	private set(implementation: ContextualImplementation): void {
		for (const concrete of this.bindingConcretes) {
			this.bindingContainer.addContextualBinding(concrete, this.dependency, implementation);
		}
	}
}

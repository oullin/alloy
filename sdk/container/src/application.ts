import { AsyncLocalStorage } from 'node:async_hooks';

import { AsyncResolutionRequiredError, MissingGlobalApplicationError, ProviderCycleError } from '#container/errors';
import { Container } from '#container/container';
import type { AnyServiceToken, ServiceToken } from '#container/token';
import type { ResolutionParameters } from '#container/types';
import type { AnyServiceProvider, AsyncServiceProvider, ProviderState, ServiceProvider } from '#container/types';

type ProviderRecord = {
	readonly provider: AnyServiceProvider;
	registered: boolean;
	deferred: boolean;
	booted: boolean;
	registering: boolean;
	booting: boolean;
	pendingRegistration?: Promise<void>;
	pendingBoot?: Promise<void>;
};

type ApplicationLifecycle = {
	readonly providersByIdentity: Map<AnyServiceProvider, ProviderRecord>;
	readonly providerOrder: ProviderRecord[];
	readonly deferredByToken: Map<AnyServiceToken, ProviderRecord>;
	hasBooted: boolean;
	bootPromise: Promise<void> | undefined;
};

const registrationContext = new AsyncLocalStorage<ProviderRecord>();
const bootContext = new AsyncLocalStorage<ProviderRecord>();

/**
 * A Container with Go-compatible provider ordering and deferred lifecycle.
 * Global helpers below are process-only; Cloudflare request composition must
 * create an Application per invocation instead.
 */
export class Application extends Container {
	private lifecycle: ApplicationLifecycle = {
		providersByIdentity: new Map<AnyServiceProvider, ProviderRecord>(),
		providerOrder: [],
		deferredByToken: new Map<AnyServiceToken, ProviderRecord>(),
		hasBooted: false,
		bootPromise: undefined,
	};

	/** Register one synchronous provider. */
	register(provider: ServiceProvider): this {
		const record = this.record(provider);

		if (record.registered || record.deferred || record.registering || record.pendingRegistration !== undefined) {
			return this;
		}

		if (this.defer(record)) {
			return this;
		}

		this.registerRecordSync(record);

		return this;
	}
	/** Register synchronous providers in Go's Kahn queue order. */
	registerMany(providers: readonly ServiceProvider[]): this {
		for (const provider of this.sortProviders(providers)) {
			this.register(provider as ServiceProvider);
		}

		return this;
	}
	/** Register a sync or async provider. */
	async registerAsync(provider: AnyServiceProvider): Promise<this> {
		const record = this.record(provider);

		if (record.registered || record.deferred) {
			if (record.registered && !record.booted && this.lifecycle.hasBooted) {
				await this.bootRecordAsync(record);
			}

			return this;
		}

		if (record.pendingRegistration !== undefined) {
			if (registrationContext.getStore() === record) {
				return this;
			}

			await record.pendingRegistration;

			return this;
		}

		if (record.registering) {
			return this;
		}

		if (this.defer(record)) {
			return this;
		}

		record.pendingRegistration = this.registerRecordAsync(record);

		try {
			await record.pendingRegistration;
		} finally {
			record.pendingRegistration = undefined;
		}

		return this;
	}
	/** Register sync or async providers in Go's Kahn queue order. */
	async registerManyAsync(providers: readonly AnyServiceProvider[]): Promise<this> {
		for (const provider of this.sortProviders(providers)) {
			await this.registerAsync(provider);
		}

		return this;
	}

	/** Boot every successfully registered synchronous provider once. */
	boot(): this {
		if (this.lifecycle.bootPromise !== undefined || this.hasPendingProviderBoot()) {
			throw new AsyncResolutionRequiredError('application boot');
		}

		const firstBoot = !this.lifecycle.hasBooted;

		if (firstBoot) {
			this.lifecycle.hasBooted = true;
		}

		for (const record of this.lifecycle.providerOrder) {
			if (record.registered && !record.booted) {
				this.bootRecordSync(record);
			}
		}

		return this;
	}
	/** Boot every successfully registered provider once, sharing concurrent work. */
	async bootAsync(): Promise<this> {
		if (this.lifecycle.bootPromise !== undefined) {
			await this.lifecycle.bootPromise;

			return this;
		}

		const needsWork = this.lifecycle.providerOrder.some((record) => record.registered && !record.booted);

		if (!needsWork && this.lifecycle.hasBooted) {
			return this;
		}

		this.lifecycle.bootPromise = (async () => {
			const firstBoot = !this.lifecycle.hasBooted;

			if (firstBoot) {
				this.lifecycle.hasBooted = true;
			}

			try {
				for (const record of this.lifecycle.providerOrder) {
					if (record.registered && !record.booted) {
						await this.bootRecordAsync(record);
					}
				}
			} catch (error) {
				if (firstBoot) {
					this.lifecycle.hasBooted = false;
				}

				throw error;
			}
		})();

		try {
			await this.lifecycle.bootPromise;
		} finally {
			this.lifecycle.bootPromise = undefined;
		}

		return this;
	}
	/** Return whether boot completed successfully. */
	booted(): boolean {
		return this.lifecycle.hasBooted;
	}
	/** Return provider lifecycle state in registration order. */
	providerStates(): readonly ProviderState[] {
		return this.lifecycle.providerOrder.map((record) => ({ provider: record.provider, registered: record.registered, deferred: record.deferred, booted: record.booted }));
	}
	/** Return all known providers, including deferred providers. */
	providers(): readonly AnyServiceProvider[] {
		return this.lifecycle.providerOrder.map((record) => record.provider);
	}
	/** Return whether any provider declares the token. */
	hasProvider(token: AnyServiceToken): boolean {
		return this.lifecycle.providerOrder.some((record) => record.provider.provides?.includes(token) ?? false);
	}
	/** Return the first provider declaring a token, matching Go ProviderFor. */
	providerFor(token: AnyServiceToken): AnyServiceProvider | undefined {
		return this.lifecycle.providerOrder.find((record) => record.provider.provides?.includes(token) ?? false)?.provider;
	}

	/** Resolve through deferred registration with empty parameters. */

	override make<T>(token: ServiceToken<T>): T {
		this.flushDeferredSync(token);

		return super.make(token);
	}
	/** Resolve through deferred registration with parameters. */

	override makeWith<T>(token: ServiceToken<T>, parameters: Readonly<Record<string, unknown>>): T {
		this.flushDeferredSync(token);

		return super.makeWith(token, parameters);
	}
	/** Resolve through async deferred registration. */

	override async makeAsync<T>(token: ServiceToken<T>): Promise<T> {
		await this.flushDeferredAsync(token);

		return await super.makeAsync(token);
	}
	/** Resolve with parameters through async deferred registration. */

	override async makeWithAsync<T>(token: ServiceToken<T>, parameters: Readonly<Record<string, unknown>>): Promise<T> {
		await this.flushDeferredAsync(token);

		return await super.makeWithAsync(token, parameters);
	}
	/** Get through deferred registration. */

	override get<T>(token: ServiceToken<T>): T {
		this.flushDeferredSync(token);

		return super.get(token);
	}
	/** Get through async deferred registration. */

	override async getAsync<T>(token: ServiceToken<T>): Promise<T> {
		await this.flushDeferredAsync(token);

		return await super.getAsync(token);
	}
	/** Create a scoped application that retains provider/deferred lifecycle composition. */
	override childScope(): Application {
		const child = new Application(this.core);

		child.lifecycle = this.lifecycle;

		return child;
	}
	/** Keep deferred resolution active for dependencies constructed by this application. */
	protected override withContext(path: readonly AnyServiceToken[], parameters: ResolutionParameters): Application {
		const child = new Application(this.core, this.scope, { path: Object.freeze([...path]), parameters });

		child.lifecycle = this.lifecycle;

		return child;
	}

	private record(provider: AnyServiceProvider): ProviderRecord {
		const existing = this.lifecycle.providersByIdentity.get(provider);

		if (existing !== undefined) {
			return existing;
		}

		const record: ProviderRecord = { provider, registered: false, deferred: false, booted: false, registering: false, booting: false };

		this.lifecycle.providersByIdentity.set(provider, record);
		this.lifecycle.providerOrder.push(record);

		return record;
	}

	private defer(record: ProviderRecord): boolean {
		if (record.provider.deferred !== true || (record.provider.provides?.length ?? 0) === 0) {
			return false;
		}

		record.deferred = true;

		for (const token of record.provider.provides ?? []) {
			this.lifecycle.deferredByToken.set(token, record);
		}

		return true;
	}

	private removeDeferred(record: ProviderRecord): void {
		for (const token of record.provider.provides ?? []) {
			if (this.lifecycle.deferredByToken.get(token) === record) {
				this.lifecycle.deferredByToken.delete(token);
			}
		}

		record.deferred = false;
	}

	private hasPendingProviderBoot(): boolean {
		return this.lifecycle.providerOrder.some((record) => record.pendingBoot !== undefined);
	}

	private registerRecordSync(record: ProviderRecord): void {
		record.registering = true;
		try {
			const provider = record.provider as ServiceProvider;

			if (provider.register === undefined) {
				throw new AsyncResolutionRequiredError('provider registration');
			}

			const result = provider.register(this);

			if (isThenable(result)) {
				throw new AsyncResolutionRequiredError('provider registration');
			}

			record.registered = true;
			if (this.lifecycle.hasBooted) {
				this.bootRecordSync(record);
			}
		} finally {
			record.registering = false;
		}
	}

	private async registerRecordAsync(record: ProviderRecord): Promise<void> {
		record.registering = true;

		try {
			const provider = record.provider as ServiceProvider & AsyncServiceProvider;

			if (provider.registerAsync !== undefined) {
				await registrationContext.run(record, async () => await provider.registerAsync(this));
			} else if (provider.register !== undefined) {
				const result = provider.register(this);

				if (isThenable(result)) {
					await result;
				}
			} else {
				throw new TypeError('container: provider must declare register or registerAsync');
			}

			record.registered = true;

			if (this.lifecycle.hasBooted) {
				await this.bootRecordAsync(record);
			}
		} finally {
			record.registering = false;
		}
	}

	private bootRecordSync(record: ProviderRecord): void {
		if (record.booted || record.booting || record.pendingBoot !== undefined) {
			if (record.pendingBoot !== undefined) {
				throw new AsyncResolutionRequiredError('provider boot');
			}

			return;
		}

		record.booting = true;

		try {
			const provider = record.provider as ServiceProvider;

			if (provider.boot !== undefined) {
				const result = provider.boot(this);

				if (isThenable(result)) {
					throw new AsyncResolutionRequiredError('provider boot');
				}
			} else if ((record.provider as AsyncServiceProvider).bootAsync !== undefined) {
				throw new AsyncResolutionRequiredError('provider boot');
			}

			record.booted = true;
		} finally {
			record.booting = false;
		}
	}

	private async bootRecordAsync(record: ProviderRecord): Promise<void> {
		if (record.booted) {
			return;
		}

		if (record.pendingBoot !== undefined) {
			if (bootContext.getStore() === record) {
				return;
			}

			await record.pendingBoot;

			return;
		}

		if (record.booting) {
			return;
		}

		record.pendingBoot = (async () => {
			record.booting = true;

			try {
				const provider = record.provider as ServiceProvider & AsyncServiceProvider;

				if (provider.bootAsync !== undefined) {
					await bootContext.run(record, async () => await provider.bootAsync?.(this));
				} else if (provider.boot !== undefined) {
					provider.boot(this);
				}

				record.booted = true;
			} finally {
				record.booting = false;
			}
		})();

		try {
			await record.pendingBoot;
		} finally {
			record.pendingBoot = undefined;
		}
	}

	private flushDeferredSync(token: AnyServiceToken): void {
		const record = this.lifecycle.deferredByToken.get(token);

		if (record === undefined || record.registered) {
			return;
		}

		if (record.pendingRegistration !== undefined || record.registering) {
			throw new AsyncResolutionRequiredError(`deferred provider for ${token.name}`);
		}

		this.removeDeferred(record);
		try {
			this.registerRecordSync(record);
		} catch (error) {
			this.defer(record);
			throw error;
		}
	}

	private async flushDeferredAsync(token: AnyServiceToken): Promise<void> {
		const record = this.lifecycle.deferredByToken.get(token);

		if (record === undefined || record.registered) {
			return;
		}

		if (record.pendingRegistration !== undefined) {
			if (registrationContext.getStore() === record) {
				return;
			}

			await record.pendingRegistration;

			return;
		}

		if (record.registering) {
			return;
		}

		this.removeDeferred(record);
		record.pendingRegistration = this.registerRecordAsync(record);

		try {
			await record.pendingRegistration;
		} catch (error) {
			this.defer(record);
			throw error;
		} finally {
			record.pendingRegistration = undefined;
		}
	}

	private sortProviders(providers: readonly AnyServiceProvider[]): readonly AnyServiceProvider[] {
		const byToken = new Map<AnyServiceToken, number>();

		for (const [index, provider] of providers.entries()) {
			for (const token of provider.provides ?? []) {
				byToken.set(token, index);
			}
		}

		const indegree = Array.from({ length: providers.length }, () => 0);
		const outgoing: number[][] = Array.from({ length: providers.length }, () => []);

		for (const [index, provider] of providers.entries()) {
			for (const token of provider.dependsOn ?? []) {
				const dependency = byToken.get(token);

				if (dependency !== undefined && dependency !== index) {
					outgoing[dependency]?.push(index);
					indegree[index] = (indegree[index] ?? 0) + 1;
				}
			}
		}

		const queue: number[] = [];

		for (let index = 0; index < providers.length; index += 1) {
			if (indegree[index] === 0) {
				queue.push(index);
			}
		}

		const ordered: AnyServiceProvider[] = [];

		while (queue.length > 0) {
			const index = queue.shift();

			if (index === undefined) {
				break;
			}

			const provider = providers[index];

			if (provider === undefined) {
				continue;
			}

			ordered.push(provider);

			for (const target of outgoing[index] ?? []) {
				indegree[target] = (indegree[target] ?? 1) - 1;
				if (indegree[target] === 0) {
					queue.push(target);
				}
			}
		}

		if (ordered.length !== providers.length) {
			const remaining = providers.filter((_, index) => (indegree[index] ?? 0) > 0);

			throw new ProviderCycleError(
				remaining,
				remaining.flatMap((provider) => provider.provides?.map((token) => token.name) ?? []),
			);
		}

		return ordered;
	}
}

const isThenable = (value: unknown): value is Promise<unknown> => typeof value === 'object' && value !== null && 'then' in value && typeof (value as { then?: unknown }).then === 'function';

let globalApplication: Application | undefined;
let globalContainer: Container | undefined;

/** Install or clear the process-wide application. */
export const setApp = (application: Application | undefined): void => {
	globalApplication = application;
};
/** Alias retained for earlier TypeScript adopters. */
export const setGlobalApplication = setApp;
/** Clear the process-wide application. */
export const clearGlobalApplication = (): void => {
	globalApplication = undefined;
};
/** Return the installed process-wide application. */
export const global = (): Application => {
	if (globalApplication === undefined) {
		throw new MissingGlobalApplicationError();
	}

	return globalApplication;
};
/** Alias retained for earlier TypeScript adopters. */
export const getGlobalApplication = global;
/** Return whether a global application exists. */
export const hasApp = (): boolean => globalApplication !== undefined;
/** Alias retained for earlier TypeScript adopters. */
export const hasGlobalApplication = hasApp;
/** Return the global application or throw. */
export const mustGlobal = global;
/** Resolve from the global application. */
export const make = <T>(token: ServiceToken<T>): T => global().make(token);
/** Throwing compatibility alias for make. */
export const mustMake = make;
/** Resolve a typed token from the global application. */
export const resolve = make;
/** Throwing compatibility alias for resolve. */
export const mustResolve = resolve;
/** Return the process-global container, creating it on first use. */

export const getInstance = (): Container => {
	globalContainer ??= new Container();

	return globalContainer;
};
/** Replace or clear the process-global container. */
export const setInstance = (container: Container | undefined): void => {
	globalContainer = container;
};

/** Go-style exported aliases for migration code. */
export {
	getInstance as GetInstance,
	setInstance as SetInstance,
	setApp as SetApp,
	global as Global,
	mustGlobal as MustGlobal,
	hasApp as HasApp,
	make as Make,
	mustMake as MustMake,
	resolve as Resolve,
	mustResolve as MustResolve,
};

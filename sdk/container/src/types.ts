import type { Application } from '#container/application';
import type { Container } from '#container/container';
import type { AnyServiceToken } from '#container/token';

/** Supported lifetime for a service binding. */
export type Lifetime = 'transient' | 'singleton' | 'scoped';
/** Parameters supplied to one parameterised factory resolution. */
export type ResolutionParameters = Readonly<Record<string, unknown>>;
/** A synchronous service factory. */
export type Factory<T> = (container: Container, parameters: ResolutionParameters) => T;
/** An asynchronous service factory. */
export type AsyncFactory<T> = (container: Container, parameters: ResolutionParameters) => Promise<T>;
/** Binding options shared by synchronous and asynchronous factories. */
export type BindingOptions = { readonly lifetime?: Lifetime };
/** A synchronous lifecycle callback. */
export type ResolvingCallback<T> = (value: T, container: Container) => void;
/** An asynchronous lifecycle callback. */
export type AsyncResolvingCallback<T> = (value: T, container: Container) => Promise<void>;
/** A synchronous extender. */
export type Extender<T> = (value: T, container: Container) => T;
/** An asynchronous extender. */
export type AsyncExtender<T> = (value: T, container: Container) => Promise<T>;
/** A synchronous method callback. */
export type MethodCallable<T> = (container: Container, parameters: ResolutionParameters) => T;
/** An asynchronous method callback. */
export type AsyncMethodCallable<T> = (container: Container, parameters: ResolutionParameters) => Promise<T>;

/** A provider accepted by synchronous application lifecycle APIs. */
export interface ServiceProvider {
	register(application: Application): void;
	boot?(application: Application): void;
	readonly provides?: readonly AnyServiceToken[];
	readonly dependsOn?: readonly AnyServiceToken[];
	readonly deferred?: boolean;
}

/** A provider accepted only by asynchronous application lifecycle APIs. */
export interface AsyncServiceProvider {
	registerAsync(application: Application): Promise<void>;
	bootAsync?(application: Application): Promise<void>;
	readonly provides?: readonly AnyServiceToken[];
	readonly dependsOn?: readonly AnyServiceToken[];
	readonly deferred?: boolean;
}

/** A lifecycle provider usable in asynchronous APIs. */
export type AnyServiceProvider = ServiceProvider | AsyncServiceProvider;
/** Snapshot returned by provider lifecycle introspection. */
export type ProviderState = { readonly provider: AnyServiceProvider; readonly registered: boolean; readonly deferred: boolean; readonly booted: boolean };

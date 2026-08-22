const anyServiceTokenBrand: unique symbol = Symbol('container.any-service-token');

/** A runtime-identity token without a service-type parameter. */
export interface AnyServiceToken {
	/** Stable human-readable name used only for diagnostics and introspection. */
	readonly name: string;
	/** @internal Nominal brand that prevents structural token forgery. */
	readonly [anyServiceTokenBrand]: true;
}

declare const serviceTokenBrand: unique symbol;

/**
 * An opaque, invariant runtime-identity service token.
 *
 * The brand is deliberately not exported. A token can only be created with
 * createServiceToken, and ServiceToken<A> is not assignable to ServiceToken<B>.
 */
export interface ServiceToken<T> extends AnyServiceToken {
	readonly [serviceTokenBrand]: (value: T) => T;
}

/** Options for creating an opaque service token. */
export type CreateServiceTokenOptions<T> = {
	/** Optional guard that enforces the otherwise-erased generic type at runtime. */
	readonly guard?: (value: unknown) => value is T;
};

const guards = new WeakMap<AnyServiceToken, (value: unknown) => boolean>();

/** Create a token with object-identity equality and a non-empty diagnostic name. */
export const createServiceToken = <T>(name: string, options: CreateServiceTokenOptions<T> = {}): ServiceToken<T> => {
	if (name.trim() === '') {
		throw new TypeError('container: service token name must not be empty');
	}

	// SAFETY: this module owns both unexported brands and freezes the only token
	// representation. Consumers cannot construct either branded interface.
	const token = Object.freeze({ name, [anyServiceTokenBrand]: true }) as ServiceToken<T>;

	if (options.guard !== undefined) {
		guards.set(token, options.guard);
	}

	return token;
};

/** @internal Return a token's optional runtime validator. */
export const serviceTokenGuard = (token: AnyServiceToken): ((value: unknown) => boolean) | undefined => guards.get(token);

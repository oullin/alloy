import type { AnyServiceToken } from '#container/token';

/** Stable public identities used by shared conformance fixtures and consumers. */
export const ContainerErrorCode = {
	MISSING_BINDING: 'MISSING_BINDING',
	CIRCULAR_RESOLUTION: 'CIRCULAR_RESOLUTION',
	SELF_ALIAS: 'SELF_ALIAS',
	ALIAS_CYCLE: 'ALIAS_CYCLE',
	INCORRECT_RESOLVED_TYPE: 'INCORRECT_RESOLVED_TYPE',
	MISSING_METHOD_BINDING: 'MISSING_METHOD_BINDING',
	ASYNC_RESOLUTION_REQUIRED: 'ASYNC_RESOLUTION_REQUIRED',
	PROVIDER_CYCLE: 'PROVIDER_CYCLE',
	MISSING_GLOBAL_APPLICATION: 'MISSING_GLOBAL_APPLICATION',
} as const;

/** A stable public error identity. */
export type ContainerErrorCode = (typeof ContainerErrorCode)[keyof typeof ContainerErrorCode];

const copiedPath = (path: readonly AnyServiceToken[]): readonly AnyServiceToken[] => Object.freeze([...path]);

/** Base class for container configuration and resolution failures. */
export class ContainerError extends Error {
	/** Stable error identity. */
	readonly code: ContainerErrorCode;

	/** Create a typed container failure. */
	protected constructor(code: ContainerErrorCode, message: string) {
		super(message);
		this.name = code;
		this.code = code;
	}
}

/** Raised when no binding, instance, or alias exists for a token. */
export class MissingBindingError extends ContainerError {
	readonly token: AnyServiceToken;

	constructor(token: AnyServiceToken) {
		super(ContainerErrorCode.MISSING_BINDING, `container: token "${token.name}" is not bound`);
		this.token = token;
	}
}

/** Raised when a resolution path depends on itself. */
export class CircularResolutionError extends ContainerError {
	readonly path: readonly AnyServiceToken[];

	constructor(path: readonly AnyServiceToken[]) {
		super(ContainerErrorCode.CIRCULAR_RESOLUTION, `container: circular resolution: ${path.map((token) => token.name).join(' -> ')}`);
		this.path = copiedPath(path);
	}
}

/** Raised when an alias points to itself. */
export class SelfAliasError extends ContainerError {
	readonly token: AnyServiceToken;

	constructor(token: AnyServiceToken) {
		super(ContainerErrorCode.SELF_ALIAS, `container: alias "${token.name}" cannot point to itself`);
		this.token = token;
	}
}

/** Raised when adding an alias would make a cycle. */
export class AliasCycleError extends ContainerError {
	readonly path: readonly AnyServiceToken[];

	constructor(path: readonly AnyServiceToken[]) {
		super(ContainerErrorCode.ALIAS_CYCLE, `container: alias cycle: ${path.map((token) => token.name).join(' -> ')}`);
		this.path = copiedPath(path);
	}
}

/** Raised when a token guard rejects a constructed value. */
export class IncorrectResolvedTypeError extends ContainerError {
	readonly token: AnyServiceToken;
	/** Safe metadata about the rejected value; the value itself is never retained. */
	readonly actualType: string;

	constructor(token: AnyServiceToken, value: unknown) {
		super(ContainerErrorCode.INCORRECT_RESOLVED_TYPE, `container: resolved value for "${token.name}" did not satisfy its token guard`);
		this.token = token;
		this.actualType = value === null ? 'null' : Array.isArray(value) ? 'array' : typeof value;
	}
}

/** Raised when a named method binding is absent. */
export class MissingMethodBindingError extends ContainerError {
	readonly method: string;

	constructor(method: string) {
		super(ContainerErrorCode.MISSING_METHOD_BINDING, `container: method binding "${method}" was not found`);
		this.method = method;
	}
}

/** Raised when synchronous resolution reaches asynchronous work. */
export class AsyncResolutionRequiredError extends ContainerError {
	readonly subject: string;

	constructor(subject: string) {
		super(ContainerErrorCode.ASYNC_RESOLUTION_REQUIRED, `container: "${subject}" requires makeAsync()`);
		this.subject = subject;
	}
}

/** Raised when provider dependencies cannot be topologically sorted. */
export class ProviderCycleError extends ContainerError {
	readonly providers: readonly object[];
	readonly path: readonly string[];

	constructor(providers: readonly object[], path: readonly string[] = []) {
		super(ContainerErrorCode.PROVIDER_CYCLE, `container: provider dependency cycle${path.length === 0 ? '' : `: ${path.join(' -> ')}`}`);
		this.providers = Object.freeze([...providers]);
		this.path = Object.freeze([...path]);
	}
}

/** Raised when a process-wide application has not been installed. */
export class MissingGlobalApplicationError extends ContainerError {
	constructor() {
		super(ContainerErrorCode.MISSING_GLOBAL_APPLICATION, 'container: no global application is installed');
	}
}

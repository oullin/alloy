export class RouteNotFoundError extends Error {
	readonly _tag = 'RouteNotFoundError';

	constructor(readonly path: string) {
		super(`route not found: ${path}`);
		this.name = 'RouteNotFoundError';
	}
}

export class MethodNotAllowedError extends Error {
	readonly _tag = 'MethodNotAllowedError';

	constructor(
		readonly path: string,
		readonly allowed: readonly string[],
	) {
		super(`the requested method is not supported for route ${path}; supported methods: ${allowed.join(', ')}`);
		this.name = 'MethodNotAllowedError';
	}
}

export class HttpResponseError extends Error {
	readonly _tag = 'HttpResponseError';

	constructor(
		readonly statusCode: number,
		message: string,
		readonly headers: Readonly<Record<string, string>> = {},
	) {
		super(`httpx: HTTP ${statusCode}: ${message}`);
		this.name = 'HttpResponseError';
	}
}

export class RouteCompilationError extends Error {
	readonly _tag = 'RouteCompilationError';

	constructor(message: string) {
		super(`route compilation error: ${message}`);
		this.name = 'RouteCompilationError';
	}
}

export class UrlGenerationError extends Error {
	readonly _tag = 'UrlGenerationError';

	constructor(message: string) {
		super(`url generation error: ${message}`);
		this.name = 'UrlGenerationError';
	}
}

export const HTTP_VERBS = ['GET', 'HEAD', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'] as const;

export type HttpVerb = (typeof HTTP_VERBS)[number];

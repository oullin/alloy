import type { CompiledRoute } from '#httpx/routing/compiled_route';

export interface MatchableRoute {
	compiled(): CompiledRoute | null;
	methods(): readonly string[];
	httpOnly(): boolean;
	secure(): boolean;
}

export interface MatchableRequest {
	method(): string;
	pathInfo(): string;
	path(): string;
	decodedPath(): string;
	host(): string;
	secure(): boolean;
	queryString?(): string;
	query?(key: string): string | undefined;
	url?(): string;
}

export interface RouteMatchingValidator {
	matches(route: MatchableRoute, request: MatchableRequest): boolean;
}

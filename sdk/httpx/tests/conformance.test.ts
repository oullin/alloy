import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vite-plus/test';
import { MethodNotAllowedError, RouteNotFoundError } from '#httpx/errors';
import { compileRoute, type SourceRoute } from '#httpx/routing/compiler';
import type { MatchableRequest } from '#httpx/routing/matching/contracts';
import { Route } from '#httpx/routing/route';
import { RouteCollection } from '#httpx/routing/route_collection';
import { UrlGenerator } from '#httpx/routing/url_generator';

class ConformanceSourceRoute implements SourceRoute {
	constructor(
		private readonly _path: string,
		private readonly _host: string = '',
		private readonly _defaults: Record<string, unknown> = {},
		private readonly _requirements: Record<string, string> = {},
	) {}

	path(): string {
		return this._path;
	}

	host(): string {
		return this._host;
	}

	requirements(): Readonly<Record<string, string>> {
		return this._requirements;
	}

	hasDefault(name: string): boolean {
		return name in this._defaults;
	}

	defaults(): Readonly<Record<string, unknown>> {
		return this._defaults;
	}
}

class ConformanceRequest implements MatchableRequest {
	constructor(
		private readonly _method: string,
		private readonly _path: string,
		private readonly _host: string = 'localhost',
		private readonly _secure: boolean = false,
	) {}

	method(): string {
		return this._method;
	}

	pathInfo(): string {
		return this._path;
	}

	path(): string {
		return this._path;
	}

	decodedPath(): string {
		return this._path;
	}

	host(): string {
		return this._host;
	}

	secure(): boolean {
		return this._secure;
	}
}

interface ConformanceCase {
	id: string;
	note: string;
	type: 'compile' | 'match' | 'generate';
	route?: {
		name?: string;
		path: string;
		host?: string;
		defaults?: Record<string, unknown>;
		requirements?: Record<string, string>;
	};
	routes?: Array<{
		method: string;
		path: string;
		name: string;
	}>;
	request?: {
		method: string;
		path: string;
		host?: string;
		secure?: boolean;
	};
	params?: Record<string, unknown>;
	expected?: {
		staticPrefix?: string;
		variables?: string[];
		name?: string;
		params?: Record<string, string>;
		url?: string;
	};
	error?: {
		code: 'ROUTE_NOT_FOUND' | 'METHOD_NOT_ALLOWED';
		allowed?: string[];
	};
}

interface ConformanceFixture {
	schemaVersion: number;
	cases: ConformanceCase[];
}

const fixturePath = fileURLToPath(new URL('../../../conformance/routing.json', import.meta.url));
const fixture: ConformanceFixture = JSON.parse(readFileSync(fixturePath, 'utf8'));

describe('Routing Conformance Fixtures', () => {
	for (const testCase of fixture.cases) {
		it(`[${testCase.type}] ${testCase.id}: ${testCase.note}`, () => {
			if (testCase.type === 'compile') {
				const r = new ConformanceSourceRoute(testCase.route!.path, testCase.route!.host ?? '', testCase.route!.defaults ?? {}, testCase.route!.requirements ?? {});

				const compiled = compileRoute(r);

				if (testCase.expected?.staticPrefix !== undefined) {
					expect(compiled.staticPrefix).toBe(testCase.expected.staticPrefix);
				}

				if (testCase.expected?.variables !== undefined) {
					expect(compiled.variables).toEqual(testCase.expected.variables);
				}
			} else if (testCase.type === 'match') {
				const collection = new RouteCollection();

				for (const def of testCase.routes ?? []) {
					const r = new Route([def.method], def.path, () => 'ok');

					r.name(def.name);
					collection.add(r);
				}

				const req = new ConformanceRequest(testCase.request!.method, testCase.request!.path, testCase.request!.host ?? 'localhost', testCase.request!.secure ?? false);

				if (testCase.error) {
					if (testCase.error.code === 'ROUTE_NOT_FOUND') {
						expect(() => collection.match(req)).toThrow(RouteNotFoundError);
					} else if (testCase.error.code === 'METHOD_NOT_ALLOWED') {
						try {
							collection.match(req);
							expect.unreachable('expected MethodNotAllowedError');
						} catch (err) {
							expect(err).toBeInstanceOf(MethodNotAllowedError);
							if (testCase.error.allowed) {
								expect([...(err as MethodNotAllowedError).allowed].sort()).toEqual([...testCase.error.allowed].sort());
							}
						}
					}
				} else if (testCase.expected) {
					const matched = collection.match(req);

					expect(matched.getName()).toBe(testCase.expected.name);
					if (testCase.expected.params) {
						expect(matched.parameters()).toEqual(testCase.expected.params);
					}
				}
			} else if (testCase.type === 'generate') {
				const collection = new RouteCollection();
				const r = new Route(['GET'], testCase.route!.path, () => 'ok');

				r.name(testCase.route!.name!);
				collection.add(r);

				const gen = new UrlGenerator(collection);
				const url = gen.route(testCase.route!.name!, testCase.params ?? {}, false);

				expect(url).toBe(testCase.expected!.url);
			}
		});
	}
});

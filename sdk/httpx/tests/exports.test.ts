import { describe, expect, it } from 'vite-plus/test';
import { HttpResponseError, MethodNotAllowedError, RouteNotFoundError } from '#httpx/errors.js';
import { HttpRequest, HttpResponse, HttpResponseFactory } from '#httpx/foundation/index.js';
import { createFetchHandler } from '#httpx/handler/index.js';
import { CorsMiddleware, RecoveryMiddleware } from '#httpx/middleware/index.js';
import { Router, UrlGenerator, routerToken, urlGeneratorToken } from '#httpx/routing/index.js';

describe('package exports', () => {
	it('exposes root errors and components', () => {
		expect(RouteNotFoundError).toBeTypeOf('function');
		expect(MethodNotAllowedError).toBeTypeOf('function');
		expect(HttpResponseError).toBeTypeOf('function');
		expect(HttpRequest).toBeTypeOf('function');
		expect(HttpResponse).toBeTypeOf('function');
		expect(HttpResponseFactory).toBeTypeOf('function');
		expect(createFetchHandler).toBeTypeOf('function');
		expect(Router).toBeTypeOf('function');
		expect(UrlGenerator).toBeTypeOf('function');
		expect(CorsMiddleware).toBeTypeOf('function');
		expect(RecoveryMiddleware).toBeTypeOf('function');
		expect(routerToken).toBeDefined();
		expect(urlGeneratorToken).toBeDefined();
	});
});

import { HttpResponseError, MethodNotAllowedError, RouteNotFoundError } from '#httpx/errors.js';
import { HttpRequest } from '#httpx/foundation/http_request.js';
import { HttpResponse } from '#httpx/foundation/http_response.js';
import type { Router } from '#httpx/routing/router.js';

export type FetchHandler = (request: Request) => Promise<Response>;

export function createFetchHandler(router: Router): FetchHandler {
	return async (request: Request): Promise<Response> => {
		try {
			const req = HttpRequest.fromFetch(request);

			const { value } = await router.dispatch(req);

			if (value instanceof Response) {
				return value;
			}

			if (value instanceof HttpResponse) {
				return value.toFetch();
			}

			if (typeof value === 'string') {
				return new Response(value, {
					status: 200,
					headers: { 'Content-Type': 'text/html; charset=utf-8' },
				});
			}

			if (value === undefined || value === null) {
				return new Response(null, { status: 204 });
			}

			if (typeof value === 'object' || typeof value === 'number' || typeof value === 'boolean') {
				return new Response(JSON.stringify(value), {
					status: 200,
					headers: { 'Content-Type': 'application/json; charset=utf-8' },
				});
			}

			return new Response(typeof value === 'symbol' || typeof value === 'bigint' ? value.toString() : String(value as string), { status: 200 });
		} catch (err) {
			if (err instanceof RouteNotFoundError) {
				return new Response(JSON.stringify({ error: err.message }), {
					status: 404,
					headers: { 'Content-Type': 'application/json; charset=utf-8' },
				});
			}

			if (err instanceof MethodNotAllowedError) {
				return new Response(JSON.stringify({ error: err.message }), {
					status: 405,
					headers: {
						'Content-Type': 'application/json; charset=utf-8',
						Allow: err.allowed.join(', '),
					},
				});
			}

			if (err instanceof HttpResponseError) {
				return new Response(JSON.stringify({ error: err.message }), {
					status: err.statusCode,
					headers: {
						'Content-Type': 'application/json; charset=utf-8',
						...err.headers,
					},
				});
			}

			return new Response(JSON.stringify({ error: 'internal server error' }), {
				status: 500,
				headers: { 'Content-Type': 'application/json; charset=utf-8' },
			});
		}
	};
}

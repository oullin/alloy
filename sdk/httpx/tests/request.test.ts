import { describe, expect, it } from 'vite-plus/test';
import { HttpRequest } from '#httpx/foundation/http_request.js';

describe('HttpRequest', () => {
	it('extracts method, path, host, scheme, and query string', () => {
		const req = HttpRequest.fromFetch(
			new Request('https://api.example.test:8443/users/profile?tab=settings&theme=dark', {
				headers: {
					Authorization: 'Bearer token-12345',
					Cookie: 'session_id=abc987; theme=dark',
					'X-Requested-With': 'XMLHttpRequest',
					Accept: 'application/json',
				},
			}),
		);

		expect(req.method()).toBe('GET');
		expect(req.isMethod('get')).toBe(true);
		expect(req.path()).toBe('/users/profile');
		expect(req.pathInfo()).toBe('/users/profile');
		expect(req.host()).toBe('api.example.test:8443');
		expect(req.scheme()).toBe('https');
		expect(req.secure()).toBe(true);
		expect(req.queryString()).toBe('tab=settings&theme=dark');
		expect(req.query('tab')).toBe('settings');
		expect(req.query('missing', 'fallback')).toBe('fallback');
		expect(req.hasQuery('tab')).toBe(true);
		expect(req.hasQuery('nope')).toBe(false);

		expect(req.bearerToken()).toBe('token-12345');
		expect(req.cookie('session_id')).toBe('abc987');
		expect(req.hasCookie('session_id')).toBe(true);
		expect(req.hasCookie('missing')).toBe(false);

		expect(req.wantsJson()).toBe(true);
		expect(req.isAjax()).toBe(true);
		expect(req.expectsJson()).toBe(true);
	});

	it('merges query and JSON body in all() with body taking precedence', async () => {
		const req = HttpRequest.fromFetch(
			new Request('https://example.test/users?page=1&limit=20&name=fromQuery', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					name: 'fromBody',
					active: true,
					count: 42,
					tags: ['admin', 'user'],
				}),
			}),
		);

		const all = await req.all();

		expect(all.page).toBe('1');
		expect(all.limit).toBe('20');
		expect(all.name).toBe('fromBody');
		expect(all.active).toBe(true);
		expect(all.count).toBe(42);

		expect(await req.input('name')).toBe('fromBody');

		expect(await req.input('missing', 'default')).toBe('default');

		const only = await req.only(['name', 'count']);

		expect(only).toEqual({ name: 'fromBody', count: 42 });

		const except = await req.except(['page', 'limit', 'tags']);

		expect(except).toEqual({ name: 'fromBody', active: true, count: 42 });

		expect(await req.boolean('active')).toBe(true);

		expect(await req.boolean('page')).toBe(true); // '1' is true

		expect(await req.integer('count')).toBe(42);

		expect(await req.integer('page')).toBe(1);
	});
});

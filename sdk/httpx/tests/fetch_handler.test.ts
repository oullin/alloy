import { describe, expect, it } from 'vite-plus/test';
import { HttpResponseError } from '#httpx/errors.js';
import { createFetchHandler } from '#httpx/handler/index.js';
import type { HttpContext } from '#httpx/routing/index.js';
import { Router } from '#httpx/routing/index.js';

describe('createFetchHandler', () => {
	it('returns JSON from matched action returning object, HttpResponse, or Response', async () => {
		const router = new Router();

		router.get('/json-obj', () => ({ message: 'hello from obj' }));
		router.get('/http-response', (ctx: HttpContext) => ctx.response.json({ message: 'hello from HttpResponse' }));
		router.get('/raw-response', () => new Response('raw response body', { status: 202 }));
		router.get('/string-html', () => '<h1>hello world</h1>');

		const handle = createFetchHandler(router);

		const res1 = await handle(new Request('https://example.test/json-obj'));

		expect(res1.status).toBe(200);

		await expect(res1.json()).resolves.toEqual({ message: 'hello from obj' });

		const res2 = await handle(new Request('https://example.test/http-response'));

		expect(res2.status).toBe(200);

		await expect(res2.json()).resolves.toEqual({ message: 'hello from HttpResponse' });

		const res3 = await handle(new Request('https://example.test/raw-response'));

		expect(res3.status).toBe(202);

		await expect(res3.text()).resolves.toBe('raw response body');

		const res4 = await handle(new Request('https://example.test/string-html'));

		expect(res4.status).toBe(200);
		expect(res4.headers.get('Content-Type')).toContain('text/html');

		await expect(res4.text()).resolves.toBe('<h1>hello world</h1>');
	});

	it('renders 404 for missing routes and 405 with Allow header for wrong method', async () => {
		const router = new Router();

		router.get('/items', () => ({ ok: true }));

		const handle = createFetchHandler(router);

		const missing = await handle(new Request('https://example.test/not-found'));

		expect(missing.status).toBe(404);

		const wrongMethod = await handle(new Request('https://example.test/items', { method: 'POST' }));

		expect(wrongMethod.status).toBe(405);
		expect(wrongMethod.headers.get('Allow')).toContain('GET');
	});

	it('renders HttpResponseError with custom status and headers', async () => {
		const router = new Router();

		router.get('/unauthorized', () => {
			throw new HttpResponseError(401, 'Unauthorized', { 'WWW-Authenticate': 'Bearer' });
		});

		const handle = createFetchHandler(router);

		const res = await handle(new Request('https://example.test/unauthorized'));

		expect(res.status).toBe(401);
		expect(res.headers.get('WWW-Authenticate')).toBe('Bearer');
	});
});

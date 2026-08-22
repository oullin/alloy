import { describe, expect, it } from 'vite-plus/test';
import { HttpResponse, HttpResponseFactory } from '#httpx/foundation/index.js';

describe('HttpResponse & HttpResponseFactory', () => {
	const factory = new HttpResponseFactory();

	it('creates json responses with headers and status', async () => {
		const res = factory.json({ ok: true, count: 5 }, 201, { 'X-Custom': 'value' });

		expect(res.statusCode()).toBe(201);
		expect(res.getHeader('x-custom')).toBe('value');
		expect(res.getHeader('content-type')).toContain('application/json');

		const fetchRes = res.toFetch();

		expect(fetchRes.status).toBe(201);
		expect(fetchRes.headers.get('Content-Type')).toContain('application/json');
		expect(fetchRes.headers.get('X-Custom')).toBe('value');

		await expect(fetchRes.json()).resolves.toEqual({ ok: true, count: 5 });
	});

	it('creates text, html, and redirect responses', async () => {
		const textRes = factory.text('hello plain', 200);
		const htmlRes = factory.html('<h1>hello</h1>', 200);
		const redirectRes = factory.redirect('/dashboard', 303);

		expect(textRes.toFetch().headers.get('Content-Type')).toContain('text/plain');
		expect(htmlRes.toFetch().headers.get('Content-Type')).toContain('text/html');
		expect(redirectRes.toFetch().status).toBe(303);
		expect(redirectRes.toFetch().headers.get('Location')).toBe('/dashboard');
	});

	it('buffers and renders cookies into Set-Cookie headers', () => {
		const res = HttpResponse.json({ ok: true });

		res.cookie({
			name: 'theme',
			value: 'dark',
			path: '/',
			httpOnly: true,
			secure: true,
			sameSite: 'lax',
		});
		res.withoutCookie('legacy_session');

		const fetchRes = res.toFetch();
		const setCookie = fetchRes.headers.get('Set-Cookie');

		expect(setCookie).toContain('theme=dark');
		expect(setCookie).toContain('HttpOnly');
		expect(setCookie).toContain('Secure');
		expect(setCookie).toContain('SameSite=lax');
		expect(setCookie).toContain('legacy_session=');
		expect(setCookie).toContain('Max-Age=-1');
	});
});

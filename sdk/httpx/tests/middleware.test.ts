import { describe, expect, it } from 'vite-plus/test';
import { HttpResponseError } from '#httpx/errors.js';
import { HttpRequest, HttpResponse } from '#httpx/foundation/index.js';

import {
	CacheHeadersMiddleware,
	CorsMiddleware,
	FrameGuardMiddleware,
	RecoveryMiddleware,
	RequestLogMiddleware,
	TrustHostsMiddleware,
	TrustProxiesMiddleware,
	ValidatePathEncodingMiddleware,
	ValidatePostSizeMiddleware,
	type Logger,
} from '#httpx/middleware/index.js';

describe('Middleware suite', () => {
	it('RecoveryMiddleware converts errors to 500 HttpResponse and reports them', async () => {
		const reported: unknown[] = [];
		const mw = new RecoveryMiddleware((err) => reported.push(err));

		const res = (await mw.handle(new Request('http://localhost'), () => {
			throw new Error('boom');
		})) as HttpResponse;

		expect(res.statusCode()).toBe(500);
		expect(reported).toHaveLength(1);
	});

	it('CorsMiddleware sets Access-Control headers and handles preflight', async () => {
		const mw = new CorsMiddleware({
			allowedOrigins: ['https://app.example.com'],
			allowedMethods: ['GET', 'POST'],
			allowedHeaders: ['Content-Type', 'Authorization'],
			allowCredentials: true,
			maxAge: 3600,
		});

		const preflightReq = HttpRequest.fromFetch(
			new Request('https://api.example.com/data', {
				method: 'OPTIONS',
				headers: {
					Origin: 'https://app.example.com',
					'Access-Control-Request-Method': 'POST',
					'Access-Control-Request-Headers': 'Content-Type',
				},
			}),
		);

		const preflightRes = (await mw.handle(preflightReq, () => HttpResponse.noContent())) as HttpResponse;

		expect(preflightRes.statusCode()).toBe(204);
		expect(preflightRes.getHeader('Access-Control-Allow-Origin')).toBe('https://app.example.com');
		expect(preflightRes.getHeader('Access-Control-Allow-Methods')).toBe('GET, POST');
		expect(preflightRes.getHeader('Access-Control-Allow-Credentials')).toBe('true');
		expect(preflightRes.getHeader('Access-Control-Max-Age')).toBe('3600');
	});

	it('FrameGuardMiddleware sets X-Frame-Options', async () => {
		const denyMw = new FrameGuardMiddleware('DENY');

		const res1 = (await denyMw.handle({}, () => HttpResponse.json({ ok: true }))) as HttpResponse;

		expect(res1.getHeader('X-Frame-Options')).toBe('DENY');

		const sameOriginMw = new FrameGuardMiddleware('SAMEORIGIN');

		const res2 = (await sameOriginMw.handle({}, () => HttpResponse.json({ ok: true }))) as HttpResponse;

		expect(res2.getHeader('X-Frame-Options')).toBe('SAMEORIGIN');
	});

	it('CacheHeadersMiddleware sets Cache-Control directives', async () => {
		const mw = new CacheHeadersMiddleware({
			public: true,
			maxAge: 3600,
			sMaxAge: 7200,
			mustRevalidate: true,
		});

		const res = (await mw.handle({}, () => HttpResponse.json({ ok: true }))) as HttpResponse;

		expect(res.getHeader('Cache-Control')).toBe('public, max-age=3600, s-maxage=7200, must-revalidate');
	});

	it('TrustHostsMiddleware validates host and rejects untrusted ones', async () => {
		const mw = new TrustHostsMiddleware('example.com', '*.example.com');

		const validReq = HttpRequest.fromFetch(new Request('https://api.example.com/test'));

		await expect(mw.handle(validReq, () => 'ok')).resolves.toBe('ok');

		const invalidReq = HttpRequest.fromFetch(new Request('https://evil.com/test'));

		await expect(mw.handle(invalidReq, () => 'ok')).rejects.toThrow(HttpResponseError);
	});

	it('TrustProxiesMiddleware strips forwarded headers when proxy is untrusted', async () => {
		const mw = new TrustProxiesMiddleware(['10.0.0.1']);

		const untrustedReq = HttpRequest.fromFetch(
			new Request('https://example.com/test', {
				headers: {
					'X-Forwarded-For': '198.51.100.1',
					'X-Real-IP': '192.168.1.50',
				},
			}),
		);

		await mw.handle(untrustedReq, () => 'ok');

		expect(untrustedReq.header('X-Forwarded-For')).toBeUndefined();
	});

	it('ValidatePathEncodingMiddleware rejects malformed URI encoding', async () => {
		const mw = new ValidatePathEncodingMiddleware();

		const validReq = HttpRequest.fromFetch(new Request('https://example.com/users/john%20doe'));

		await expect(mw.handle(validReq, () => 'ok')).resolves.toBe('ok');

		const invalidReq = HttpRequest.fromFetch(new Request('https://example.com/users/bad%2'));

		await expect(mw.handle(invalidReq, () => 'ok')).rejects.toThrow(HttpResponseError);
	});

	it('ValidatePostSizeMiddleware rejects bodies over limit', async () => {
		const mw = new ValidatePostSizeMiddleware(1024);

		const validReq = HttpRequest.fromFetch(
			new Request('https://example.com/upload', {
				method: 'POST',
				headers: { 'Content-Length': '500' },
			}),
		);

		await expect(mw.handle(validReq, () => 'ok')).resolves.toBe('ok');

		const oversizedReq = HttpRequest.fromFetch(
			new Request('https://example.com/upload', {
				method: 'POST',
				headers: { 'Content-Length': '2048' },
			}),
		);

		await expect(mw.handle(oversizedReq, () => 'ok')).rejects.toThrow(HttpResponseError);
	});

	it('RequestLogMiddleware logs request details using injected logger', async () => {
		const logs: Array<{ message: string; context?: Record<string, unknown> }> = [];

		const logger: Logger = {
			info: (msg, ctx) => logs.push({ message: msg, context: ctx }),
			error: (msg, ctx) => logs.push({ message: msg, context: ctx }),
		};

		const mw = new RequestLogMiddleware({ logger, skipPaths: ['/health'] });
		const req1 = HttpRequest.fromFetch(new Request('https://example.com/api/v1/users'));

		await mw.handle(req1, () => HttpResponse.json({ ok: true }));

		expect(logs).toHaveLength(1);
		expect(logs[0].message).toBe('http: request');
		expect(logs[0].context?.method).toBe('GET');
		expect(logs[0].context?.path).toBe('/api/v1/users');
		expect(logs[0].context?.status).toBe(200);

		const req2 = HttpRequest.fromFetch(new Request('https://example.com/health'));

		await mw.handle(req2, () => HttpResponse.json({ ok: true }));

		expect(logs).toHaveLength(1); // skipped
	});
});

import { describe, expect, it } from 'vite-plus/test';
import { UrlGenerationError } from '#httpx/errors';
import { Route } from '#httpx/routing/route';
import { RouteCollection } from '#httpx/routing/route_collection';
import { UrlGenerator, type URLRequest } from '#httpx/routing/url_generator';

class TestUrlRequest implements URLRequest {
	constructor(
		private readonly _url: string,
		private readonly _path: string,
		private readonly _host: string = 'localhost',
		private readonly _scheme: string = 'http',
	) {}

	url(): string {
		return this._url;
	}

	path(): string {
		return this._path;
	}

	host(): string {
		return this._host;
	}

	scheme(): string {
		return this._scheme;
	}

	queryString(): string {
		const idx = this._url.indexOf('?');

		return idx >= 0 ? this._url.slice(idx + 1) : '';
	}

	query(key: string): string | undefined {
		const qs = this.queryString();

		if (!qs) {
			return undefined;
		}

		const sp = new URLSearchParams(qs);

		return sp.get(key) ?? undefined;
	}
}

describe('UrlGenerator', () => {
	it('generates path and asset URLs', () => {
		const routes = new RouteCollection();
		const gen = new UrlGenerator(routes);

		gen.forceRootUrl('http://localhost:3000');

		expect(gen.to('/users')).toBe('http://localhost:3000/users');
		expect(gen.to('/users', ['123', 'profile'])).toBe('http://localhost:3000/users/123/profile');
		expect(gen.secure('/login')).toBe('https://localhost:3000/login');
		expect(gen.asset('css/app.css')).toBe('http://localhost:3000/css/app.css');
	});

	it('generates named route URLs with parameters and query strings', () => {
		const routes = new RouteCollection();
		const r = new Route(['GET'], '/users/{id}/posts/{post?}', () => 'ok');

		r.name('users.posts.show');
		routes.add(r);

		const gen = new UrlGenerator(routes);

		gen.forceRootUrl('http://localhost:3000');

		expect(gen.route('users.posts.show', { id: 5, post: 'hello' })).toBe('http://localhost:3000/users/5/posts/hello');
		expect(gen.route('users.posts.show', { id: 5 })).toBe('http://localhost:3000/users/5/posts');
		expect(gen.route('users.posts.show', { id: 5, page: 2, sort: 'desc' })).toBe('http://localhost:3000/users/5/posts?page=2&sort=desc');
		expect(gen.route('users.posts.show', { id: 5 }, false)).toBe('/users/5/posts');
	});

	it('throws on missing required parameters for named routes', () => {
		const routes = new RouteCollection();
		const r = new Route(['GET'], '/users/{id}', () => 'ok');

		r.name('users.show');
		routes.add(r);

		const gen = new UrlGenerator(routes);

		expect(() => gen.route('users.show', {})).toThrow(UrlGenerationError);
	});

	it('signs and validates URLs using HMAC-SHA256', () => {
		const routes = new RouteCollection();
		const r = new Route(['GET'], '/verify-email/{id}', () => 'ok');

		r.name('email.verify');
		routes.add(r);

		const gen = new UrlGenerator(routes);

		gen.forceRootUrl('http://localhost:3000');
		gen.setKeyResolver('super-secret-key-1234567890');

		const signedUrl = gen.signedRoute('email.verify', { id: 42 });

		expect(signedUrl).toContain('signature=');

		const urlObj = new URL(signedUrl);

		const req = new TestUrlRequest(signedUrl, urlObj.pathname, urlObj.host, urlObj.protocol.replace(':', ''));

		expect(gen.hasValidSignature(req)).toBe(true);

		// Tampered parameter should fail
		const tamperedReq = new TestUrlRequest(signedUrl.replace('/42', '/43'), '/verify-email/43', urlObj.host, urlObj.protocol.replace(':', ''));

		expect(gen.hasValidSignature(tamperedReq)).toBe(false);
	});

	it('handles expiration for temporary signed URLs', () => {
		const routes = new RouteCollection();
		const r = new Route(['GET'], '/download/{token}', () => 'ok');

		r.name('file.download');
		routes.add(r);

		const gen = new UrlGenerator(routes);

		gen.forceRootUrl('http://localhost:3000');
		gen.setKeyResolver('super-secret-key-1234567890');

		// Valid future expiration (60 seconds)
		const validUrl = gen.temporarySignedRoute('file.download', 60, { token: 'abc' });
		const validObj = new URL(validUrl);

		const validReq = new TestUrlRequest(validUrl, validObj.pathname, validObj.host, validObj.protocol.replace(':', ''));

		expect(gen.hasValidSignature(validReq)).toBe(true);

		// Expired timestamp (in the past)
		const expiredUrl = gen.signedRoute('file.download', { token: 'abc' }, -60);
		const expiredObj = new URL(expiredUrl);

		const expiredReq = new TestUrlRequest(expiredUrl, expiredObj.pathname, expiredObj.host, expiredObj.protocol.replace(':', ''));

		expect(gen.hasValidSignature(expiredReq)).toBe(false);
	});
});

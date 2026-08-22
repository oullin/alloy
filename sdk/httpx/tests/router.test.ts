import { describe, expect, it } from 'vite-plus/test';
import { MethodNotAllowedError, RouteNotFoundError } from '#httpx/errors';
import type { MatchableRequest } from '#httpx/routing/matching/contracts';
import { Router } from '#httpx/routing/router';

class TestRequest implements MatchableRequest {
	constructor(
		private readonly _method: string,
		private readonly _path: string,
		private readonly _host = 'localhost',
		private readonly _secure = false,
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

describe('Router', () => {
	it('registers and matches basic verb routes', async () => {
		const router = new Router();

		router.get('/users', () => 'users index');
		router.post('/users', () => 'user created');
		router.get('/users/{id}', (_req, params) => `user ${(params as { id: string }).id}`);

		const getRes = await router.dispatch(new TestRequest('GET', '/users'));

		expect(getRes.value).toBe('users index');

		const postRes = await router.dispatch(new TestRequest('POST', '/users'));

		expect(postRes.value).toBe('user created');

		const userRes = await router.dispatch(new TestRequest('GET', '/users/42'));

		expect(userRes.value).toBe('user 42');
	});

	it('supports route groups with prefix, as, and middleware', async () => {
		const router = new Router();
		const events: string[] = [];

		router.aliasMiddleware('auth', async (_req: unknown, next: (req?: unknown) => unknown) => {
			events.push('auth');

			return next(_req);
		});

		router.group({ prefix: 'admin', as: 'admin.', middleware: 'auth' }, (groupRouter) => {
			groupRouter
				.get('/dashboard', () => {
					events.push('dashboard');

					return 'admin dashboard';
				})
				.name('dashboard');
		});

		expect(router.has('admin.dashboard')).toBe(true);

		const route = router.getRoutes().getByName('admin.dashboard');

		expect(route?.uri).toBe('/admin/dashboard');

		const res = await router.dispatch(new TestRequest('GET', '/admin/dashboard'));

		expect(res.value).toBe('admin dashboard');
		expect(events).toEqual(['auth', 'dashboard']);
	});

	it('supports resource routing', () => {
		const router = new Router();

		router.resource('photos', {
			index: () => 'index',
			create: () => 'create',
			store: () => 'store',
			show: () => 'show',
			edit: () => 'edit',
			update: () => 'update',
			destroy: () => 'destroy',
		});

		expect(router.has('photos.index')).toBe(true);
		expect(router.has('photos.create')).toBe(true);
		expect(router.has('photos.store')).toBe(true);
		expect(router.has('photos.show')).toBe(true);
		expect(router.has('photos.edit')).toBe(true);
		expect(router.has('photos.update')).toBe(true);
		expect(router.has('photos.destroy')).toBe(true);

		const showRoute = router.getRoutes().getByName('photos.show');

		expect(showRoute?.uri).toBe('/photos/{photo}');
	});

	it('supports apiResource routing (drops create and edit)', () => {
		const router = new Router();

		router.apiResource('posts', {});

		expect(router.has('posts.index')).toBe(true);
		expect(router.has('posts.store')).toBe(true);
		expect(router.has('posts.show')).toBe(true);
		expect(router.has('posts.update')).toBe(true);
		expect(router.has('posts.destroy')).toBe(true);
		expect(router.has('posts.create')).toBe(false);
		expect(router.has('posts.edit')).toBe(false);
	});

	it('throws RouteNotFoundError on unmatched routes', async () => {
		const router = new Router();

		router.get('/hello', () => 'world');

		await expect(router.dispatch(new TestRequest('GET', '/nonexistent'))).rejects.toThrow(RouteNotFoundError);
	});

	it('throws MethodNotAllowedError when path matches different method', async () => {
		const router = new Router();

		router.post('/items', () => 'created');

		await expect(router.dispatch(new TestRequest('GET', '/items'))).rejects.toThrow(MethodNotAllowedError);
	});

	it('supports fallback routes', async () => {
		const router = new Router();

		router.get('/home', () => 'home');
		router.fallback(() => 'fallback catchall');

		const res = await router.dispatch(new TestRequest('GET', '/anything-else'));

		expect(res.value).toBe('fallback catchall');
	});
});

import { describe, expect, it, vi } from 'vite-plus/test';

import { createRouteResolver, fillPattern, resolveRoute, resolveRouteUrl } from '#expose-routes/index';

describe('expose manifest routing', () => {
	it('fills route parameters from a manifest', () => {
		const route = resolveRoute(
			{
				'contacts.show': '/contacts/{contact}',
			},
			'contacts.show',
			{ contact: 42 },
			{ onMissingRoute: false },
		);

		expect(route).toEqual({ url: '/contacts/42' });
	});

	it('encodes route parameter values', () => {
		expect(fillPattern('/files/{path}', { path: 'invoices/June 2026.pdf' })).toBe('/files/invoices%2FJune%202026.pdf');
	});

	it('removes missing optional route segments', () => {
		expect(fillPattern('/archive/{year?}/{month?}', { year: 2026 })).toBe('/archive/2026');
	});

	it('supports binding field placeholders', () => {
		expect(fillPattern('/posts/{post:slug}', { post: 'hello-world' })).toBe('/posts/hello-world');
	});

	it('returns a fallback and reports missing routes', () => {
		const onMissingRoute = vi.fn();

		const url = resolveRouteUrl({}, 'missing.route', {}, { onMissingRoute });

		expect(url).toBe('#!expose:unknown-route');
		expect(onMissingRoute).toHaveBeenCalledWith('missing.route', '#!expose:unknown-route');
	});

	it('resolves empty root route patterns', () => {
		const onMissingRoute = vi.fn();

		const url = resolveRouteUrl({ home: '' }, 'home', {}, { onMissingRoute });

		expect(url).toBe('');
		expect(onMissingRoute).not.toHaveBeenCalled();
	});

	it('creates reusable resolvers from dynamic route manifests', () => {
		const routes = { dashboard: '/dashboard' };
		const route = createRouteResolver(() => routes, { onMissingRoute: false });

		expect(route('dashboard')).toEqual({ url: '/dashboard' });
	});
});

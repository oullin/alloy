import { describe, expect, it, vi } from 'vite-plus/test';

import { createRouteResolver, fillPattern, resolveRoute, resolveRouteUrl } from '#navigator-routes/index';
import type { PatternParams } from '#navigator-routes/index';

// Widens a literal pattern back to `string`. Leaving a required parameter
// unfilled is a type error on a literal pattern, so the runtime behaviour for
// unfilled parameters is only reachable the way manifest-driven callers reach
// it: through a pattern the compiler cannot see into.
const widened = (pattern: string): string => pattern;

describe('navigator manifest routing', () => {
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

describe('colon route parameters', () => {
	it('fills express and hono style parameters', () => {
		expect(fillPattern('/users/:user/posts/:post', { user: 7, post: 'hello world' })).toBe('/users/7/posts/hello%20world');
	});

	it('mixes colon and brace parameters in one pattern', () => {
		expect(fillPattern('/teams/:team/members/{member}/{tab?}', { team: 'core', member: 3 })).toBe('/teams/core/members/3');
	});

	it('leaves unfilled colon parameters in place outside strict mode', () => {
		expect(fillPattern(widened('/users/:user'), {})).toBe('/users/:user');
	});

	it('never treats a protocol colon as a parameter', () => {
		expect(fillPattern('https://cdn.example.com/assets/:asset', { asset: 'logo.svg' })).toBe('https://cdn.example.com/assets/logo.svg');
		expect(fillPattern('https://cdn.example.com/assets', {})).toBe('https://cdn.example.com/assets');
	});

	it('never treats a binding constraint as a colon parameter', () => {
		expect(fillPattern(widened('/posts/{post:slug}'), {})).toBe('/posts/{post:slug}');
		expect(fillPattern(widened('/posts/{post:slug}'), { slug: 'ignored' })).toBe('/posts/{post:slug}');
	});

	it('resolves colon patterns through a manifest', () => {
		const route = createRouteResolver({ 'posts.show': '/posts/:post' }, { onMissingRoute: false });

		expect(route('posts.show', { post: 42 })).toEqual({ url: '/posts/42' });
	});
});

describe('strict mode', () => {
	it('throws when a required brace parameter is unfilled', () => {
		expect(() => fillPattern(widened('/posts/{post}'), {}, { strict: true })).toThrow('missing required parameter "post" for route pattern "/posts/{post}"');
	});

	it('throws when a required colon parameter is unfilled', () => {
		expect(() => fillPattern(widened('/users/:user/posts'), {}, { strict: true })).toThrow('missing required parameter "user" for route pattern "/users/:user/posts"');
	});

	it('throws when a required constrained parameter is unfilled', () => {
		expect(() => fillPattern(widened('/posts/{post:slug}'), {}, { strict: true })).toThrow('missing required parameter "post" for route pattern "/posts/{post:slug}"');
	});

	it('collapses optional segments without throwing', () => {
		expect(fillPattern('/archive/{year?}/{month?}', { year: 2026 }, { strict: true })).toBe('/archive/2026');
		expect(fillPattern('/archive/{year?}', {}, { strict: true })).toBe('/archive');
	});

	it('throws instead of filling a missing pattern', () => {
		expect(() => fillPattern(null, {}, { strict: true })).toThrow('cannot fill a missing route pattern');
	});

	it('throws on unknown routes instead of returning the sentinel', () => {
		const onMissingRoute = vi.fn();

		expect(() => resolveRouteUrl({}, 'missing.route', {}, { strict: true, onMissingRoute })).toThrow('unknown route "missing.route"');
		expect(() => resolveRoute({}, 'missing.route', {}, { strict: true, onMissingRoute })).toThrow('unknown route "missing.route"');
		expect(onMissingRoute).not.toHaveBeenCalled();
	});

	it('throws from resolvers built over a manifest', () => {
		const route = createRouteResolver({ dashboard: '/dashboard' }, { strict: true });

		expect(route('dashboard')).toEqual({ url: '/dashboard' });
		expect(() => route('missing')).toThrow('unknown route "missing"');
	});

	it('leaves the non-strict default untouched', () => {
		expect(fillPattern(widened('/posts/{post}'), {})).toBe('/posts/{post}');
		expect(resolveRouteUrl({}, 'missing.route', {}, { onMissingRoute: false })).toBe('#!expose:unknown-route');
		expect(fillPattern(widened('/posts/{post}'), {}, { strict: false })).toBe('/posts/{post}');
	});
});

describe('pattern parameter types', () => {
	// The `@ts-expect-error` assertions below are type-level: they are armed by
	// `vp check`, not by the vitest run, which only executes the statements.
	it('requires every parameter a literal pattern names', () => {
		expect(fillPattern('/users/:user/posts/{post}', { user: 7, post: 3 })).toBe('/users/7/posts/3');

		// @ts-expect-error - `post` is a required parameter of this literal pattern.
		expect(fillPattern('/users/:user/posts/{post}', { user: 7 })).toBe('/users/7/posts/{post}');

		// @ts-expect-error - `slug` is not a parameter of this literal pattern.
		expect(fillPattern('/users/:user', { user: 7, slug: 'nope' })).toBe('/users/7');
	});

	it('treats brace constraints as plain required parameters', () => {
		expect(fillPattern('/posts/{post:slug}', { post: 'hello-world' })).toBe('/posts/hello-world');

		// @ts-expect-error - the `:slug` constraint is not part of the parameter name.
		expect(fillPattern('/posts/{post:slug}', { 'post:slug': 'hello-world' })).toBe('/posts/{post:slug}');
	});

	it('keeps optional parameters optional', () => {
		expect(fillPattern('/archive/{year?}/{month?}')).toBe('/archive');
		expect(fillPattern('/archive/{year?}/{month?}', { month: 6 })).toBe('/archive/6');

		// @ts-expect-error - `day` is not a parameter of this literal pattern.
		expect(fillPattern('/archive/{year?}', { day: 1 })).toBe('/archive');
	});

	it('exposes the parameter shape through PatternParams', () => {
		const params: PatternParams<'/teams/:team/members/{member}/{tab?}'> = { team: 'core', member: 3 };

		expect(fillPattern('/teams/:team/members/{member}/{tab?}', params)).toBe('/teams/core/members/3');

		// @ts-expect-error - `member` is required by this pattern.
		const incomplete: PatternParams<'/teams/:team/members/{member}'> = { team: 'core' };

		expect(incomplete).toEqual({ team: 'core' });
	});

	it('falls back to RouteParams for widened string patterns', () => {
		const dynamic: string = '/posts/{post}';
		const params: PatternParams<string> = { post: 42 };

		expect(fillPattern(dynamic, params)).toBe('/posts/42');

		// @ts-expect-error - widened patterns still restrict values to RouteParamValue.
		expect(fillPattern(dynamic, { post: { id: 42 } })).toBe('/posts/%5Bobject%20Object%5D');
	});

	it('accepts null and undefined patterns unchanged', () => {
		expect(fillPattern(null, {}, { onMissingRoute: false })).toBe('#!expose:unknown-route');
		expect(fillPattern(undefined)).toBe('#!expose:unknown-route');
	});
});

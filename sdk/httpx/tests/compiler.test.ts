import { describe, expect, it } from 'vite-plus/test';
import { RouteCompilationError } from '#httpx/errors';
import { compileRoute, type SourceRoute } from '#httpx/routing/compiler';

class FakeRoute implements SourceRoute {
	constructor(
		private readonly _path: string,
		private readonly _host: string = '',
		private readonly _defaults: Record<string, unknown> = {},
		private readonly _requirements: Record<string, string> = {},
	) {}

	path(): string {
		return this._path;
	}

	host(): string {
		return this._host;
	}

	requirements(): Readonly<Record<string, string>> {
		return this._requirements;
	}

	hasDefault(name: string): boolean {
		return name in this._defaults;
	}

	defaults(): Readonly<Record<string, unknown>> {
		return this._defaults;
	}
}

describe('RouteCompiler', () => {
	it('compiles static path', () => {
		const c = compileRoute(new FakeRoute('/foo'));

		expect(c.staticPrefix).toBe('/foo');
		expect(c.compiledRegex.test('/foo')).toBe(true);
		expect(c.compiledRegex.test('/bar')).toBe(false);
	});

	it('compiles single path variable', () => {
		const c = compileRoute(new FakeRoute('/users/{user}'));

		expect(c.pathVariables).toEqual(['user']);

		const m = c.compiledRegex.exec('/users/42');

		expect(m).not.toBeNull();
		expect(m?.groups?.['user']).toBe('42');
	});

	it('compiles optional variable with default', () => {
		const c = compileRoute(new FakeRoute('/users/{user}', '', { user: 'guest' }));

		expect(c.compiledRegex.test('/users')).toBe(true);
		expect(c.compiledRegex.test('/users/alice')).toBe(true);

		const m = c.compiledRegex.exec('/users/alice');

		expect(m?.groups?.['user']).toBe('alice');
	});

	it('compiles variable requirements', () => {
		const c = compileRoute(new FakeRoute('/users/{id}', '', {}, { id: '[0-9]+' }));

		expect(c.compiledRegex.test('/users/123')).toBe(true);
		expect(c.compiledRegex.test('/users/abc')).toBe(false);
	});

	it('compiles host patterns', () => {
		const c = compileRoute(new FakeRoute('/', '{sub}.example.com'));

		expect(c.compiledHostRegex).not.toBeNull();

		const m = c.compiledHostRegex?.exec('api.example.com');

		expect(m).not.toBeNull();
		expect(m?.groups?.['sub']).toBe('api');
	});

	it('throws on duplicate variable names', () => {
		expect(() => compileRoute(new FakeRoute('/{x}/{x}'))).toThrow(RouteCompilationError);
	});

	it('throws on digit-leading variable names', () => {
		expect(() => compileRoute(new FakeRoute('/{1bad}'))).toThrow(RouteCompilationError);
	});

	it('throws on variable names longer than 32 chars', () => {
		expect(() => compileRoute(new FakeRoute('/{aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa}'))).toThrow(RouteCompilationError);
	});

	it('throws on _fragment as path variable', () => {
		expect(() => compileRoute(new FakeRoute('/{_fragment}'))).toThrow(RouteCompilationError);
	});

	it('compiles two variables with intermediate text', () => {
		const c = compileRoute(new FakeRoute('/posts/{post}/comments/{comment}'));
		const m = c.compiledRegex.exec('/posts/1/comments/2');

		expect(m).not.toBeNull();
		expect(m?.groups?.['post']).toBe('1');
		expect(m?.groups?.['comment']).toBe('2');
	});
});

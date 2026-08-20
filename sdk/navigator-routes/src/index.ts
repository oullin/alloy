export type RouteManifest = Record<string, string | null | undefined>;

export type RouteParamValue = string | number | boolean;

export type RouteParams = Record<string, RouteParamValue | null | undefined>;

export type RouteResult = {
	url: string;
};

export type MissingRouteReporter = (name: string, fallback: string) => void;

export type NavigatorOptions = {
	fallback?: string;
	onMissingRoute?: MissingRouteReporter | false;
	/**
	 * Fail loudly instead of falling back. Unknown route names throw instead of
	 * returning the sentinel URL, and unfilled required parameters throw instead
	 * of being passed through verbatim. Optional `{param?}` segments still
	 * collapse. Defaults to `false` for backwards compatibility.
	 */
	strict?: boolean;
};

/** @deprecated Use NavigatorOptions. */
export type ExposeOptions = NavigatorOptions;

export type RouteResolver = (name: string, params?: RouteParams) => RouteResult;

type LowercaseLetter = 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z';

type Digit = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9';

/** Mirrors the `[A-Za-z0-9_]` token grammar used by the `:param` matcher. */
type ColonParamChar = Digit | LowercaseLetter | Uppercase<LowercaseLetter> | '_';

/** Splits the longest `[A-Za-z0-9_]` prefix off a string as `[name, rest]`. */
type ColonParamName<TSource extends string, TName extends string = ''> = TSource extends `${infer Char}${infer Rest}`
	? Char extends ColonParamChar
		? ColonParamName<Rest, `${TName}${Char}`>
		: [TName, TSource]
	: [TName, TSource];

/** Every `:param` name in a brace-free slice of a pattern. `https://` yields nothing. */
type ColonParamKeys<TSource extends string> = TSource extends `${string}:${infer Rest}`
	? ColonParamName<Rest> extends [infer TName extends string, infer TRest extends string]
		? (TName extends '' ? never : TName) | ColonParamKeys<TRest>
		: never
	: never;

/** Drops the `:regex` constraint from a `{param:regex}` token body. */
type BraceParamName<TToken extends string> = TToken extends `${infer TName}:${string}` ? TName : TToken;

type PatternRequiredKeys<TPattern extends string> = TPattern extends `${infer THead}{${infer TToken}}${infer TTail}`
	? ColonParamKeys<THead> | (TToken extends `${string}?` ? never : BraceParamName<TToken>) | PatternRequiredKeys<TTail>
	: ColonParamKeys<TPattern>;

type PatternOptionalKeys<TPattern extends string> = TPattern extends `${string}{${infer TToken}}${infer TTail}`
	? (TToken extends `${infer TName}?` ? BraceParamName<TName> : never) | PatternOptionalKeys<TTail>
	: never;

type Simplify<T> = { [K in keyof T]: T[K] };

/**
 * The parameters a literal route pattern needs: `{param}` and `:param` are
 * required, `{param?}` is optional, and `{param:regex}` contributes `param`.
 * Widened `string` patterns fall back to the untyped `RouteParams` record.
 */
export type PatternParams<TPattern extends string> = string extends TPattern
	? RouteParams
	: Simplify<{ [TKey in PatternRequiredKeys<TPattern>]: RouteParamValue } & { [TKey in PatternOptionalKeys<TPattern>]?: RouteParamValue }>;

/** The trailing `fillPattern` arguments; `params` is required only when the pattern names one. */
export type FillPatternArgs<TPattern extends string | null | undefined> = [TPattern] extends [string]
	? string extends TPattern
		? [params?: RouteParams, options?: NavigatorOptions]
		: [PatternRequiredKeys<TPattern & string>] extends [never]
			? [params?: PatternParams<TPattern & string>, options?: NavigatorOptions]
			: [params: PatternParams<TPattern & string>, options?: NavigatorOptions]
	: [params?: RouteParams, options?: NavigatorOptions];

const defaultFallback = '#!expose:unknown-route';
const errorPrefix = '[navigator-routes]';
// One pass, three alternatives: an optional `/{param?}` segment, a `{param}`
// token (with an optional `:regex` constraint and `?` marker), then a bare
// `:param` token. Brace tokens are matched first so a `{param:regex}`
// constraint is never mistaken for a `:param`, and `[A-Za-z0-9_]` excludes `/`
// so `https://` never matches.
const routeTokenRe = /\/\{([A-Za-z0-9_-]+)(?::[^}?]+)?\?\}|\{([A-Za-z0-9_-]+)(?::[^}?]+)?(\?)?\}|:([A-Za-z0-9_]+)/g;
const hasOwn = <T extends object>(value: T, key: PropertyKey): key is keyof T => Object.prototype.hasOwnProperty.call(value, key);
const routeFallback = (options?: NavigatorOptions): string => options?.fallback ?? defaultFallback;

const encodeRouteParam = (value: RouteParamValue): string => encodeURIComponent(String(value));

const defaultReporter: MissingRouteReporter = (name) => {
	console.warn(`[expose] unknown route "${name}", returning fallback`);
};

const reportMissing = (name: string, fallback: string, options?: NavigatorOptions): void => {
	if (options?.onMissingRoute === false) {
		return;
	}

	const reporter = options?.onMissingRoute ?? defaultReporter;

	reporter(name, fallback);
};

const normalizedUrl = (url: string): string => {
	const normalized = url.replace(/([^:])\/{2,}/g, '$1/');

	if (normalized.length > 1 && normalized.endsWith('/')) {
		return normalized.slice(0, -1);
	}

	return normalized;
};

const missingParamError = (pattern: string, key: string): Error => new Error(`${errorPrefix} missing required parameter "${key}" for route pattern "${pattern}"`);

const unknownRouteError = (name: string): Error => new Error(`${errorPrefix} unknown route "${name}"`);

const fillRoutePattern = (pattern: string | null | undefined, params: RouteParams, options?: NavigatorOptions): string => {
	const strict = options?.strict === true;

	if (strict && (pattern === undefined || pattern === null)) {
		throw new Error(`${errorPrefix} cannot fill a missing route pattern`);
	}

	const source = pattern ?? routeFallback(options);

	const url = source.replace(routeTokenRe, (token: string, optionalKey?: string, braceKey?: string, optionalMark?: string, colonKey?: string): string => {
		if (optionalKey !== undefined) {
			const optionalValue = params[optionalKey];

			return optionalValue === undefined || optionalValue === null ? '' : `/${encodeRouteParam(optionalValue)}`;
		}

		const key = braceKey ?? colonKey;

		if (key === undefined) {
			return token;
		}

		const value = params[key];

		if (value !== undefined && value !== null) {
			return encodeRouteParam(value);
		}

		if (optionalMark !== undefined) {
			return strict ? '' : token;
		}

		if (strict) {
			throw missingParamError(source, key);
		}

		return token;
	});

	return normalizedUrl(url);
};

export const fillPattern = <const TPattern extends string | null | undefined>(pattern: TPattern, ...args: FillPatternArgs<TPattern>): string => {
	const [params, options] = args as [RouteParams | undefined, NavigatorOptions | undefined];

	return fillRoutePattern(pattern, params ?? {}, options);
};

export const resolveRouteUrl = (manifest: RouteManifest | null | undefined, name: string, params: RouteParams = {}, options?: NavigatorOptions): string => {
	const pattern = manifest && hasOwn(manifest, name) ? manifest[name] : undefined;

	if (pattern === undefined || pattern === null) {
		if (options?.strict === true) {
			throw unknownRouteError(name);
		}

		const fallback = routeFallback(options);

		reportMissing(name, fallback, options);

		return fallback;
	}

	return fillRoutePattern(pattern, params, options);
};

export const resolveRoute = (manifest: RouteManifest | null | undefined, name: string, params: RouteParams = {}, options?: NavigatorOptions): RouteResult => ({
	url: resolveRouteUrl(manifest, name, params, options),
});

export const createRouteResolver = (manifest: RouteManifest | (() => RouteManifest | null | undefined), options?: NavigatorOptions): RouteResolver => {
	return (name, params = {}) => {
		const routes = typeof manifest === 'function' ? manifest() : manifest;

		return resolveRoute(routes, name, params, options);
	};
};

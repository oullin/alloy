export type RouteManifest = Record<string, string | null | undefined>;

export type RouteParamValue = string | number | boolean;

export type RouteParams = Record<string, RouteParamValue | null | undefined>;

export type RouteResult = {
	url: string;
};

export type MissingRouteReporter = (name: string, fallback: string) => void;

export type ExposeOptions = {
	fallback?: string;
	onMissingRoute?: MissingRouteReporter | false;
};

export type RouteResolver = (name: string, params?: RouteParams) => RouteResult;

const defaultFallback = '#!expose:unknown-route';
const optionalSegmentRe = /\/\{([A-Za-z0-9_-]+)(?::[^}?]+)?\?\}/g;
const parameterRe = /\{([A-Za-z0-9_-]+)(?::[^}?]+)?\??\}/g;
const hasOwn = <T extends object>(value: T, key: PropertyKey): key is keyof T => Object.prototype.hasOwnProperty.call(value, key);
const routeFallback = (options?: ExposeOptions): string => options?.fallback ?? defaultFallback;
const encodeRouteParam = (value: RouteParamValue): string => encodeURIComponent(String(value));

const defaultReporter: MissingRouteReporter = (name) => {
	console.warn(`[expose] unknown route "${name}", returning fallback`);
};

const reportMissing = (name: string, fallback: string, options?: ExposeOptions): void => {
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

export const fillPattern = (pattern: string | null | undefined, params: RouteParams = {}, options?: ExposeOptions): string => {
	let url = pattern ?? routeFallback(options);

	url = url.replace(optionalSegmentRe, (_segment, key: string) => {
		const value = params[key];

		if (value === undefined || value === null) {
			return '';
		}

		return `/${encodeRouteParam(value)}`;
	});

	url = url.replace(parameterRe, (placeholder, key: string) => {
		const value = params[key];

		if (value === undefined || value === null) {
			return placeholder;
		}

		return encodeRouteParam(value);
	});

	return normalizedUrl(url);
};

export const resolveRouteUrl = (manifest: RouteManifest | null | undefined, name: string, params: RouteParams = {}, options?: ExposeOptions): string => {
	const pattern = manifest && hasOwn(manifest, name) ? manifest[name] : undefined;

	if (!pattern) {
		const fallback = routeFallback(options);

		reportMissing(name, fallback, options);

		return fallback;
	}

	return fillPattern(pattern, params, options);
};

export const resolveRoute = (manifest: RouteManifest | null | undefined, name: string, params: RouteParams = {}, options?: ExposeOptions): RouteResult => ({
	url: resolveRouteUrl(manifest, name, params, options),
});

export const createRouteResolver = (manifest: RouteManifest | (() => RouteManifest | null | undefined), options?: ExposeOptions): RouteResolver => {
	return (name, params = {}) => {
		const routes = typeof manifest === 'function' ? manifest() : manifest;

		return resolveRoute(routes, name, params, options);
	};
};

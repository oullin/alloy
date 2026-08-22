import { createHmac, timingSafeEqual } from 'node:crypto';
import { UrlGenerationError } from '#httpx/errors';
import type { Route } from '#httpx/routing/route';
import type { RouteCollection } from '#httpx/routing/route_collection';

export interface URLRequest {
	url(): string;
	path(): string;
	host(): string;
	scheme(): string;
	queryString(): string;
	query(key: string): string | undefined;
}

export class UrlGenerator {
	private _key = '';
	private _forcedScheme = '';
	private _forcedRootUrl = '';
	private _assetRoot = '';

	constructor(
		private readonly routes: RouteCollection,
		private readonly request?: URLRequest,
		assetRoot = '',
	) {
		this._assetRoot = assetRoot;
	}

	setKeyResolver(key: string | (() => string)): this {
		this._key = typeof key === 'function' ? key() : key;

		return this;
	}

	forceScheme(scheme: string): this {
		this._forcedScheme = scheme;

		return this;
	}

	forceHttps(force = true): this {
		this._forcedScheme = force ? 'https' : '';

		return this;
	}

	forceRootUrl(root: string): this {
		this._forcedRootUrl = root;

		return this;
	}

	full(): string {
		return this.request ? this.request.url() : '';
	}

	current(): string {
		if (!this.request) {
			return '';
		}

		const full = this.request.url();
		const idx = full.indexOf('?');

		return idx >= 0 ? full.slice(0, idx) : full;
	}

	to(path: string, extra: string[] = [], secure?: boolean): string {
		if (isAbsoluteUrl(path)) {
			return path;
		}

		const root = this.formatRoot(this.formatScheme(secure), '');

		let tail = path.replace(/^\/+/u, '');

		for (const e of extra) {
			tail = `${tail.replace(/\/+$/u, '')}/${encodeURIComponent(e)}`;
		}

		return `${root}/${tail}`;
	}

	secure(path: string, extra: string[] = []): string {
		return this.to(path, extra, true);
	}

	asset(path: string, secure?: boolean): string {
		if (isAbsoluteUrl(path)) {
			return path;
		}

		const root = this._assetRoot !== '' ? this._assetRoot : this.formatRoot(this.formatScheme(secure), '');

		return `${root.replace(/\/+$/u, '')}/${path.replace(/^\/+/u, '')}`;
	}

	route(name: string, parameters: Record<string, unknown> = {}, absolute = true): string {
		const route = this.routes.getByName(name);

		if (!route) {
			throw new UrlGenerationError(`route [${name}] not defined`);
		}

		return this.toRoute(route, parameters, absolute);
	}

	toRoute(route: Route, parameters: Record<string, unknown> = {}, absolute = true): string {
		const merged = this.mergeRouteDefaults(route, parameters);

		for (const name of route.parameterNames()) {
			if (!(name in merged) || merged[name] === undefined) {
				if (!route.hasDefault(name) && !this.isOptionalParam(route.uri, name)) {
					throw new UrlGenerationError(`missing required parameter for [Route: ${route.getName()}] [URI: ${route.uri}] [Missing parameter: ${name}]`);
				}
			}
		}

		return this.generateRouteUrl(route, merged, absolute);
	}

	signedRoute(name: string, parameters: Record<string, unknown> = {}, expiration?: Date | number, absolute = true): string {
		if ('signature' in parameters) {
			throw new UrlGenerationError('"signature" is a reserved parameter for signed routes');
		}

		if ('expires' in parameters) {
			throw new UrlGenerationError('"expires" is a reserved parameter for signed routes');
		}

		const params = { ...parameters };

		if (expiration !== undefined && expiration !== null) {
			const expSeconds =
				expiration instanceof Date ? Math.floor(expiration.getTime() / 1000) : typeof expiration === 'number' && expiration !== 0 ? Math.floor(Date.now() / 1000) + expiration : 0;

			if (expSeconds !== 0) {
				params.expires = String(expSeconds);
			}
		}

		const base = this.route(name, params, absolute);
		const mac = createHmac('sha256', this._key);

		mac.update(base);

		const signature = mac.digest('hex');

		params.signature = signature;

		return this.route(name, params, absolute);
	}

	temporarySignedRoute(name: string, expiration: Date | number, parameters: Record<string, unknown> = {}, absolute = true): string {
		return this.signedRoute(name, parameters, expiration, absolute);
	}

	hasValidSignature(request: URLRequest, absolute = true): boolean {
		return this.hasCorrectSignature(request, absolute) && this.signatureHasNotExpired(request);
	}

	hasValidRelativeSignature(request: URLRequest): boolean {
		return this.hasValidSignature(request, false);
	}

	hasCorrectSignature(request: URLRequest, absolute = true): boolean {
		let urlStr = request.url();

		if (!absolute) {
			urlStr = `/${request.path().replace(/^\/+/u, '')}`;
		}

		const qIdx = urlStr.indexOf('?');

		if (qIdx >= 0) {
			urlStr = urlStr.slice(0, qIdx);
		}

		const queryString = this.canonicalizeQuery(request.queryString());

		let original = urlStr;

		if (queryString !== '') {
			original += `?${queryString}`;
		}

		const expected = request.query('signature') ?? '';
		const mac = createHmac('sha256', this._key);

		mac.update(original);

		const got = mac.digest('hex');

		if (got.length !== expected.length) {
			return false;
		}

		try {
			return timingSafeEqual(Buffer.from(got, 'utf8'), Buffer.from(expected, 'utf8'));
		} catch {
			return false;
		}
	}

	signatureHasNotExpired(request: URLRequest): boolean {
		const expires = request.query('expires');

		if (!expires) {
			return true;
		}

		const exp = parseInt(expires, 10);

		if (Number.isNaN(exp)) {
			return false;
		}

		return Math.floor(Date.now() / 1000) <= exp;
	}

	private generateRouteUrl(route: Route, parameters: Record<string, unknown>, absolute: boolean): string {
		const domain = route.getDomain();

		let uri = `/${route.uri.replace(/^\/+/u, '')}`;

		const consumed = new Set<string>();

		uri = uri.replace(/\{([\w]+)\??\}/g, (_match, name) => {
			if (name in parameters && parameters[name] !== undefined) {
				consumed.add(name);

				const val = parameters[name];
				const strVal = typeof val === 'string' ? val : String(val as number | boolean);

				return encodeURIComponent(strVal);
			}

			return _match;
		});

		// Strip unconsumed optional placeholders /{param?}
		uri = uri.replace(/\/\{[\w]+\?\}/g, '');

		const queryPairs: Array<[string, string]> = [];
		const sortedKeys = Object.keys(parameters).sort();

		for (const k of sortedKeys) {
			if (!consumed.has(k) && parameters[k] !== undefined && parameters[k] !== null) {
				const val = parameters[k];
				const strVal = typeof val === 'string' ? val : String(val as number | boolean);

				queryPairs.push([k, strVal]);
			}
		}

		if (queryPairs.length > 0) {
			const searchParams = new URLSearchParams();

			for (const [k, v] of queryPairs) {
				searchParams.append(k, v);
			}

			uri += `?${searchParams.toString()}`;
		}

		if (!absolute) {
			return uri;
		}

		const root = this.formatRoot(this.formatScheme(route.secure() ? true : undefined), domain);

		return `${root}${uri}`;
	}

	private canonicalizeQuery(rawQuery: string): string {
		if (rawQuery === '') {
			return '';
		}

		const searchParams = new URLSearchParams(rawQuery);

		searchParams.delete('signature');

		searchParams.sort();

		return searchParams.toString();
	}

	private mergeRouteDefaults(route: Route, parameters: Record<string, unknown>): Record<string, unknown> {
		return { ...route.defaultValues, ...parameters };
	}

	private isOptionalParam(uri: string, name: string): boolean {
		return uri.includes(`{${name}?}`);
	}

	private formatScheme(secure?: boolean): string {
		if (this._forcedScheme !== '') {
			return this._forcedScheme;
		}

		if (secure !== undefined) {
			return secure ? 'https' : 'http';
		}

		if (this.request) {
			return this.request.scheme();
		}

		return 'http';
	}

	private formatRoot(scheme: string, root: string): string {
		if (this._forcedRootUrl !== '') {
			root = this._forcedRootUrl;
		}

		if (root === '' && this.request) {
			root = `${scheme}://${this.request.host()}`;
		}

		if (root === '') {
			root = `${scheme}://`;
		}

		if (root.includes('://')) {
			root = `${scheme}${root.slice(root.indexOf('://'))}`;
		} else {
			root = `${scheme}://${root}`;
		}

		return root.replace(/\/+$/u, '');
	}
}

function isAbsoluteUrl(path: string): boolean {
	return path.startsWith('http://') || path.startsWith('https://') || path.startsWith('//') || path.startsWith('mailto:') || path.startsWith('tel:');
}

import type { MatchableRequest } from '#httpx/routing/matching/contracts.js';

export interface SessionStore {
	get(key: string, fallback?: unknown): unknown;
	put(key: string, value: unknown): void;
	flash(key: string, value: unknown): void;
	getOldInput(key: string, fallback?: unknown): unknown;
	hasOldInput(key: string): boolean;
	flashInput(values: Record<string, unknown>): void;
	remove(key: string): unknown;
}

export interface RouteResolver {
	currentRouteName(): string | undefined;
	currentRouteAction(): unknown;
}

export class HttpRequest implements MatchableRequest {
	private _session?: SessionStore;
	private _routeResolver?: RouteResolver;
	private _parsedInput: Record<string, unknown> | null = null;
	private readonly _url: URL;

	constructor(private readonly _raw: Request) {
		this._url = new URL(_raw.url);
	}

	static fromFetch(req: Request): HttpRequest {
		return new HttpRequest(req);
	}

	static fromMatchable(matchable: MatchableRequest): HttpRequest {
		const scheme = matchable.secure() ? 'https' : 'http';
		const host = matchable.host() || 'localhost';
		const path = matchable.pathInfo().startsWith('/') ? matchable.pathInfo() : `/${matchable.pathInfo()}`;
		const url = `${scheme}://${host}${path}`;
		const raw = new Request(url, { method: matchable.method() });

		return new HttpRequest(raw);
	}

	raw(): Request {
		return this._raw;
	}

	method(): string {
		return this._raw.method;
	}

	isMethod(method: string): boolean {
		return this._raw.method.toLowerCase() === method.toLowerCase();
	}

	url(): string {
		return this._raw.url;
	}

	fullUrl(): string {
		return this._url.toString();
	}

	path(): string {
		return this._url.pathname;
	}

	pathInfo(): string {
		return this._url.pathname;
	}

	decodedPath(): string {
		try {
			return decodeURIComponent(this._url.pathname);
		} catch {
			return this._url.pathname;
		}
	}

	scheme(): string {
		return this._url.protocol.replace(':', '');
	}

	secure(): boolean {
		return this._url.protocol === 'https:';
	}

	host(): string {
		return this._url.host;
	}

	port(): string {
		return this._url.port;
	}

	ip(): string {
		return this.header('x-forwarded-for')?.split(',')[0]?.trim() ?? this.header('x-real-ip') ?? '127.0.0.1';
	}

	queryString(): string {
		return this._url.search.startsWith('?') ? this._url.search.slice(1) : this._url.search;
	}

	query(key: string, fallback?: string): string | undefined {
		const val = this._url.searchParams.get(key);

		return val !== null ? val : fallback;
	}

	hasQuery(key: string): boolean {
		return this._url.searchParams.has(key);
	}

	header(key: string, fallback?: string): string | undefined {
		const val = this._raw.headers.get(key);

		return val !== null ? val : fallback;
	}

	hasHeader(key: string): boolean {
		return this._raw.headers.has(key);
	}

	headers(): Headers {
		return this._raw.headers;
	}

	bearerToken(): string | undefined {
		const auth = this.header('authorization');

		if (auth && auth.toLowerCase().startsWith('bearer ')) {
			return auth.slice(7).trim();
		}

		return undefined;
	}

	cookie(name: string, fallback?: string): string | undefined {
		const cookieHeader = this.header('cookie');

		if (!cookieHeader) {
			return fallback;
		}

		const cookies = cookieHeader.split(';');

		for (const cookie of cookies) {
			const [k, ...v] = cookie.trim().split('=');

			if (k === name) {
				return decodeURIComponent(v.join('='));
			}
		}

		return fallback;
	}

	hasCookie(name: string): boolean {
		return this.cookie(name) !== undefined;
	}

	setSession(session: SessionStore): void {
		this._session = session;
	}

	session(): SessionStore | undefined {
		return this._session;
	}

	setRouteResolver(resolver: RouteResolver): void {
		this._routeResolver = resolver;
	}

	routeResolver(): RouteResolver | undefined {
		return this._routeResolver;
	}

	wantsJson(): boolean {
		const acceptable = this.header('accept') ?? '';

		return acceptable.includes('application/json') || acceptable.includes('application/vnd.api+json') || acceptable.includes('+json');
	}

	accepts(contentType: string): boolean {
		const acceptable = this.header('accept') ?? '';

		return acceptable.includes(contentType) || acceptable.includes('*/*');
	}

	expectsJson(): boolean {
		return this.wantsJson() || this.isAjax();
	}

	isJson(): boolean {
		const ct = this.header('content-type') ?? '';

		return ct.includes('/json') || ct.includes('+json');
	}

	isAjax(): boolean {
		return this.header('x-requested-with')?.toLowerCase() === 'xmlhttprequest';
	}

	async all(): Promise<Record<string, unknown>> {
		if (this._parsedInput !== null) {
			return this._parsedInput;
		}

		const merged: Record<string, unknown> = {};

		// Query params first
		for (const [k, v] of this._url.searchParams.entries()) {
			merged[k] = v;
		}

		// Body input
		const contentType = this.header('content-type') ?? '';

		if (this.isMethod('GET') || this.isMethod('HEAD')) {
			this._parsedInput = merged;

			return merged;
		}

		try {
			if (contentType.includes('application/json')) {
				const cloned = this._raw.clone();

				const json = await cloned.json();

				if (typeof json === 'object' && json !== null) {
					Object.assign(merged, json);
				}
			} else if (contentType.includes('application/x-www-form-urlencoded') || contentType.includes('multipart/form-data')) {
				const cloned = this._raw.clone();

				const formData = await cloned.formData();

				for (const [k, v] of formData.entries()) {
					merged[k] = v;
				}
			}
		} catch {
			// Body reading failure or malformed JSON
		}

		this._parsedInput = merged;

		return merged;
	}

	async input<T = unknown>(key: string, fallback?: T): Promise<T> {
		const all = await this.all();

		return (all[key] as T) ?? (fallback as T);
	}

	async only(keys: readonly string[]): Promise<Record<string, unknown>> {
		const all = await this.all();

		const result: Record<string, unknown> = {};

		for (const key of keys) {
			if (key in all) {
				result[key] = all[key];
			}
		}

		return result;
	}

	async except(keys: readonly string[]): Promise<Record<string, unknown>> {
		const all = await this.all();

		const result: Record<string, unknown> = { ...all };

		for (const key of keys) {
			delete result[key];
		}

		return result;
	}

	async boolean(key: string, fallback = false): Promise<boolean> {
		const val = await this.input(key);

		if (val === undefined || val === null) {
			return fallback;
		}

		if (typeof val === 'boolean') {
			return val;
		}

		const str = (typeof val === 'string' ? val : String(val as number | boolean)).toLowerCase().trim();

		return str === '1' || str === 'true' || str === 'on' || str === 'yes';
	}

	async integer(key: string, fallback = 0): Promise<number> {
		const val = await this.input(key);

		if (val === undefined || val === null) {
			return fallback;
		}

		const parsed = typeof val === 'number' ? Math.trunc(val) : parseInt(typeof val === 'string' ? val : String(val as boolean), 10);

		return Number.isNaN(parsed) ? fallback : parsed;
	}
}

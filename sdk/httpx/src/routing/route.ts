import type { CompiledRoute } from '#httpx/routing/compiled_route';
import { compileRoute, type SourceRoute } from '#httpx/routing/compiler';

import { HostValidator, MethodValidator, SchemeValidator, UriValidator, type MatchableRequest, type MatchableRoute, type RouteMatchingValidator } from '#httpx/routing/matching/index';

export type RouteHandler = (context?: any, ...args: any[]) => unknown;

export interface RouteActionObject {
	uses?: unknown;
	handler?: string;
	middleware?: unknown[];
	where?: Record<string, string>;
	defaults?: Record<string, unknown>;
	domain?: string;
	prefix?: string;
	as?: string;
	[key: string]: unknown;
}

export type RouteAction = RouteHandler | RouteActionObject | string;

export class Route implements SourceRoute, MatchableRoute {
	private _uri: string;
	private _methods: string[];
	private _action: RouteActionObject;
	private _wheres: Record<string, string> = {};
	private _defaults: Record<string, unknown> = {};
	private _parameters: Record<string, string> | null = null;
	private _bindingFields: Record<string, string> = {};
	private _isFallback = false;
	private _isHttpOnly = false;
	private _isSecure = false;
	private _compiled: CompiledRoute | null = null;
	private _middleware: unknown[] = [];
	private _router: unknown = null;
	private _validators: RouteMatchingValidator[] | null = null;

	constructor(methods: string | readonly string[], uri: string, action?: RouteAction) {
		this._uri = uri;
		this._methods = normalizeMethods(methods);

		if (typeof action === 'function') {
			this._action = { uses: action };
		} else if (typeof action === 'object' && action !== null) {
			this._action = { ...action };
		} else if (typeof action === 'string') {
			this._action = { uses: action, handler: action };
		} else {
			this._action = {};
		}

		// Symfony / HTTP convention: append HEAD when route handles GET
		const hasGet = this._methods.includes('GET');
		const hasHead = this._methods.includes('HEAD');

		if (hasGet && !hasHead) {
			this._methods.push('HEAD');
		}

		if (this._action.prefix) {
			this.prefix(this._action.prefix);
		}

		if (this._action.domain) {
			this.domain(this._action.domain);
		}

		if (this._action.middleware) {
			this.middleware(this._action.middleware);
		}

		if (this._action.where) {
			this.where(this._action.where);
		}

		if (this._action.defaults) {
			for (const [k, v] of Object.entries(this._action.defaults)) {
				this.defaults(k, v);
			}
		}

		const parsedUri = parseRouteUri(this._uri);

		this._uri = parsedUri.uri;

		for (const [k, v] of Object.entries(parsedUri.bindingFields)) {
			this._bindingFields[k] = v;
		}
	}

	get uri(): string {
		return this._uri;
	}

	get isFallback(): boolean {
		return this._isFallback;
	}

	get bindingFields(): Readonly<Record<string, string>> {
		return this._bindingFields;
	}

	get defaultValues(): Readonly<Record<string, unknown>> {
		return this._defaults;
	}

	get actionMap(): Readonly<RouteActionObject> {
		return this._action;
	}

	get action(): Readonly<RouteActionObject> {
		return this._action;
	}

	methods(): readonly string[] {
		return this._methods;
	}

	path(): string {
		return this._uri;
	}

	host(): string {
		return (this._action.domain as string) ?? '';
	}

	requirements(): Readonly<Record<string, string>> {
		return this._wheres;
	}

	hasDefault(name: string): boolean {
		return name in this._defaults;
	}

	defaults(): Readonly<Record<string, unknown>>;

	defaults(key: string, value: unknown): this;

	defaults(key?: string, value?: unknown): this | Readonly<Record<string, unknown>> {
		if (key === undefined) {
			return { ...this._defaults };
		}

		this._defaults[key] = value;
		this._compiled = null;

		return this;
	}

	setDefaults(defaults: Record<string, unknown>): this {
		this._defaults = { ...this._defaults, ...defaults };
		this._compiled = null;

		return this;
	}

	getDefault(key: string): unknown {
		return this._defaults[key];
	}

	name(name: string): this {
		if (this._action.as) {
			if (!this._action.as.endsWith(name)) {
				this._action.as = `${this._action.as}${name}`;
			}
		} else {
			this._action.as = name;
		}

		return this;
	}

	getName(): string {
		return (this._action.as as string) ?? '';
	}

	prefix(prefix: string): this {
		const cleanPrefix = prefix.replace(/\/+$/u, '').replace(/^\/+/u, '');
		const cleanUri = this._uri.replace(/^\/+/u, '');

		if (cleanPrefix !== '') {
			this._uri = `/${cleanPrefix}/${cleanUri}`.replace(/\/+$/u, '');
			if (this._uri === '') {
				this._uri = '/';
			}
		} else if (!this._uri.startsWith('/')) {
			this._uri = `/${cleanUri}`;
		}

		this._compiled = null;

		return this;
	}

	getPrefix(): string {
		return (this._action.prefix as string) ?? '';
	}

	domain(domain: string): this {
		this._action.domain = domain;
		this._compiled = null;

		return this;
	}

	getDomain(): string {
		return (this._action.domain as string) ?? '';
	}

	hasDomain(): boolean {
		return typeof this._action.domain === 'string' && this._action.domain !== '';
	}

	middleware(middleware: unknown): this {
		if (Array.isArray(middleware)) {
			this._middleware.push(...middleware);
		} else if (middleware !== undefined && middleware !== null) {
			this._middleware.push(middleware);
		}

		return this;
	}

	gatherMiddleware(): unknown[] {
		return [...this._middleware];
	}

	fallback(): this {
		this._isFallback = true;

		return this;
	}

	httpOnly(): boolean {
		return this._isHttpOnly;
	}

	setHttpOnly(flag = true): this {
		this._isHttpOnly = flag;
		if (flag) {
			this._isSecure = false;
		}

		return this;
	}

	secure(): boolean {
		return this._isSecure;
	}

	setSecure(flag = true): this {
		this._isSecure = flag;
		if (flag) {
			this._isHttpOnly = false;
		}

		return this;
	}

	where(nameOrObj: string | Record<string, string>, expression?: string): this {
		if (typeof nameOrObj === 'object' && nameOrObj !== null) {
			for (const [k, v] of Object.entries(nameOrObj)) {
				this._wheres[k] = v;
			}
		} else if (typeof nameOrObj === 'string' && expression !== undefined) {
			this._wheres[nameOrObj] = expression;
		}

		this._compiled = null;

		return this;
	}

	whereAlpha(name: string): this {
		return this.where(name, '[a-zA-Z]+');
	}

	whereAlphaNumeric(name: string): this {
		return this.where(name, '[a-zA-Z0-9]+');
	}

	whereNumber(name: string): this {
		return this.where(name, '[0-9]+');
	}

	whereUuid(name: string): this {
		return this.where(name, '[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}');
	}

	whereUlid(name: string): this {
		return this.where(name, '[0-7][0-9A-HJKMNP-TV-Z]{25}');
	}

	whereIn(name: string, values: readonly string[]): this {
		const escaped = values.map((v) => v.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'));

		return this.where(name, escaped.join('|'));
	}

	compile(): CompiledRoute {
		if (this._compiled === null) {
			this._compiled = compileRoute(this);
		}

		return this._compiled;
	}

	compiled(): CompiledRoute | null {
		return this.compile();
	}

	parameterNames(): readonly string[] {
		return this.compile().variables;
	}

	parameters(): Readonly<Record<string, string>> {
		return this._parameters ?? {};
	}

	parameter(name: string, defaultValue?: string): string | undefined {
		return this._parameters?.[name] ?? defaultValue;
	}

	matches(request: MatchableRequest, includingMethod = true): boolean {
		this.compile();

		const validators = this.getValidators();

		for (const validator of validators) {
			if (!includingMethod && validator instanceof MethodValidator) {
				continue;
			}

			if (!validator.matches(this, request)) {
				return false;
			}
		}

		return true;
	}

	bind(request: MatchableRequest): this {
		const compiled = this.compile();
		const params: Record<string, string> = {};

		let path = request.decodedPath();

		if (!path.startsWith('/')) {
			path = `/${path}`;
		}

		const match = compiled.compiledRegex.exec(path);

		if (match?.groups) {
			for (const [k, v] of Object.entries(match.groups)) {
				if (v !== undefined) {
					params[k] = v;
				}
			}
		}

		if (compiled.compiledHostRegex !== null) {
			const hostMatch = compiled.compiledHostRegex.exec(request.host());

			if (hostMatch?.groups) {
				for (const [k, v] of Object.entries(hostMatch.groups)) {
					if (v !== undefined && !(k in params)) {
						params[k] = v;
					}
				}
			}
		}

		for (const name of compiled.variables) {
			if (!(name in params) || params[name] === undefined) {
				if (this.hasDefault(name)) {
					const def = this._defaults[name];

					params[name] = def !== null && def !== undefined ? (typeof def === 'string' ? def : String(def as string | number | boolean)) : '';
				}
			}
		}

		this._parameters = params;

		return this;
	}

	cloneForRequest(): Route {
		const r = new Route(this._methods, this._uri, { ...this._action });

		r._wheres = { ...this._wheres };
		r._defaults = { ...this._defaults };
		r._bindingFields = { ...this._bindingFields };
		r._isFallback = this._isFallback;
		r._isHttpOnly = this._isHttpOnly;
		r._isSecure = this._isSecure;
		r._middleware = [...this._middleware];
		r._router = this._router;

		return r;
	}

	async run(request?: unknown): Promise<unknown> {
		const uses = this._action.uses;

		if (typeof uses === 'function') {
			return uses(request, this._parameters);
		}

		if (typeof uses === 'object' && uses !== null && 'invoke' in uses && typeof (uses as { invoke: RouteHandler }).invoke === 'function') {
			return (uses as { invoke: RouteHandler }).invoke(request, this._parameters);
		}

		return uses;
	}

	private getValidators(): RouteMatchingValidator[] {
		if (this._validators === null) {
			this._validators = [new UriValidator(), new MethodValidator(), new HostValidator(), new SchemeValidator()];
		}

		return this._validators;
	}
}

function normalizeMethods(methods: string | readonly string[]): string[] {
	const list = Array.isArray(methods) ? methods : [methods];

	return list.map((m) => m.toUpperCase());
}

interface ParsedRouteUri {
	uri: string;
	bindingFields: Record<string, string>;
}

function parseRouteUri(uri: string): ParsedRouteUri {
	const bindingFields: Record<string, string> = {};

	const cleaned = uri.replace(/\{([\w]+):([\w]+)\}/g, (_match, param, field) => {
		bindingFields[param] = field;

		return `{${param}}`;
	});

	let normalized = `/${cleaned.replace(/^\/+/u, '')}`;

	if (normalized.length > 1) {
		normalized = normalized.replace(/\/+$/u, '');
	}

	return { uri: normalized, bindingFields };
}

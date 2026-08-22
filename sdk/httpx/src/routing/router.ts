import type { Container } from '@hara/sdk-container';
import { HTTP_VERBS } from '#httpx/errors.js';
import { HttpRequest } from '#httpx/foundation/http_request.js';
import { HttpResponse } from '#httpx/foundation/http_response.js';
import { HttpResponseFactory } from '#httpx/foundation/http_response_factory.js';
import { HttpContext } from './http_context.js';
import type { MatchableRequest } from './matching/contracts.js';
import { ResourceRegistrar, type ResourceOptions } from './resource_registrar.js';
import { Route, type RouteAction } from './route.js';
import { RouteCollection } from './route_collection.js';
import { mergeRouteGroup } from './route_group.js';

export interface DispatchResult {
	route: Route;
	value: unknown;
	context: HttpContext;
}

export type MiddlewareHandler = (request: unknown, next: (req?: unknown) => unknown, ...args: unknown[]) => unknown;

export interface MiddlewareObject {
	handle(request: unknown, next: (req?: unknown) => unknown, ...args: unknown[]): unknown;
}

export class RouteGroupBuilder {
	constructor(
		private readonly router: Router,
		private readonly attributes: Record<string, unknown> = {},
	) {}

	prefix(prefix: string): this {
		this.attributes.prefix = prefix;

		return this;
	}

	domain(domain: string): this {
		this.attributes.domain = domain;

		return this;
	}

	as(name: string): this {
		this.attributes.as = name;

		return this;
	}

	name(name: string): this {
		return this.as(name);
	}

	middleware(middleware: unknown): this {
		this.attributes.middleware = middleware;

		return this;
	}

	where(where: Record<string, string>): this {
		this.attributes.where = where;

		return this;
	}

	group(callback: (router: Router) => void): Router {
		return this.router.group(this.attributes, callback);
	}
}

export class Router {
	private readonly _routes = new RouteCollection();
	private readonly _middleware = new Map<string, unknown>();
	private readonly _middlewareGroups = new Map<string, unknown[]>();
	private readonly _patterns: Record<string, string> = {};
	private readonly _groupStack: Array<Record<string, unknown>> = [];
	public middlewarePriority: string[] = [];

	constructor(private _container?: Container) {}

	get container(): Container | undefined {
		return this._container;
	}

	setContainer(container: Container): this {
		this._container = container;

		return this;
	}

	get routes(): RouteCollection {
		return this._routes;
	}

	getRoutes(): RouteCollection {
		return this._routes;
	}

	addRoute(methods: string | readonly string[], uri: string, action?: RouteAction): Route {
		const route = this.createRoute(methods, uri, action);

		return this._routes.add(route);
	}

	get(uri: string, action?: RouteAction): Route {
		return this.addRoute(['GET', 'HEAD'], uri, action);
	}

	post(uri: string, action?: RouteAction): Route {
		return this.addRoute('POST', uri, action);
	}

	put(uri: string, action?: RouteAction): Route {
		return this.addRoute('PUT', uri, action);
	}

	patch(uri: string, action?: RouteAction): Route {
		return this.addRoute('PATCH', uri, action);
	}

	delete(uri: string, action?: RouteAction): Route {
		return this.addRoute('DELETE', uri, action);
	}

	options(uri: string, action?: RouteAction): Route {
		return this.addRoute('OPTIONS', uri, action);
	}

	any(uri: string, action?: RouteAction): Route {
		return this.addRoute(HTTP_VERBS, uri, action);
	}

	match(methods: string | readonly string[], uri: string, action?: RouteAction): Route {
		return this.addRoute(methods, uri, action);
	}

	fallback(action: RouteAction): Route {
		const placeholder = 'fallbackPlaceholder';

		return this.addRoute('GET', `/{${placeholder}}`, action).where(placeholder, '.*').fallback();
	}

	redirect(uri: string, destination: string, status = 302): Route {
		return this.any(uri, (ctx) => {
			if (ctx instanceof HttpContext) {
				return ctx.response.redirect(destination, status);
			}

			return HttpResponse.redirect(destination, status);
		});
	}

	permanentRedirect(uri: string, destination: string): Route {
		return this.redirect(uri, destination, 301);
	}

	prefix(prefix: string): RouteGroupBuilder {
		return new RouteGroupBuilder(this).prefix(prefix);
	}

	domain(domain: string): RouteGroupBuilder {
		return new RouteGroupBuilder(this).domain(domain);
	}

	as(name: string): RouteGroupBuilder {
		return new RouteGroupBuilder(this).as(name);
	}

	name(name: string): RouteGroupBuilder {
		return this.as(name);
	}

	middleware(middleware: unknown): RouteGroupBuilder {
		return new RouteGroupBuilder(this).middleware(middleware);
	}

	group(attributes: Record<string, unknown>, callback: (router: Router) => void): this {
		this.updateGroupStack(attributes);
		try {
			callback(this);
		} finally {
			this._groupStack.pop();
		}

		return this;
	}

	resource(name: string, controller: unknown, options: ResourceOptions = {}): RouteCollection {
		return new ResourceRegistrar(this).register(name, controller, options);
	}

	apiResource(name: string, controller: unknown, options: ResourceOptions = {}): RouteCollection {
		return new ResourceRegistrar(this).register(name, controller, {
			...options,
			except: ['create', 'edit', ...(options.except ?? [])],
		});
	}

	singleton(name: string, controller: unknown, options: ResourceOptions = {}): RouteCollection {
		return new ResourceRegistrar(this).singleton(name, controller, options);
	}

	apiSingleton(name: string, controller: unknown, options: ResourceOptions = {}): RouteCollection {
		return new ResourceRegistrar(this).singleton(name, controller, {
			...options,
			except: ['edit', ...(options.except ?? [])],
		});
	}

	pattern(key: string, pattern: string): this {
		this._patterns[key] = pattern;

		return this;
	}

	patterns(patterns: Record<string, string>): this {
		for (const [k, v] of Object.entries(patterns)) {
			this._patterns[k] = v;
		}

		return this;
	}

	aliasMiddleware(name: string, middleware: unknown): this {
		this._middleware.set(name, middleware);

		return this;
	}

	middlewareGroup(name: string, middleware: unknown[]): this {
		this._middlewareGroups.set(name, middleware);

		return this;
	}

	getMiddlewareGroups(): ReadonlyMap<string, unknown[]> {
		return this._middlewareGroups;
	}

	getMiddleware(): ReadonlyMap<string, unknown> {
		return this._middleware;
	}

	has(name: string): boolean {
		return this._routes.hasNamedRoute(name);
	}

	current(): Route | undefined {
		// TypeScript router avoids process-wide mutable state for concurrent dispatch safety
		return undefined;
	}

	currentRouteName(): string | undefined {
		return undefined;
	}

	currentRouteAction(): unknown {
		return undefined;
	}

	async dispatch(request: MatchableRequest | HttpRequest): Promise<DispatchResult> {
		const req = request instanceof HttpRequest ? request : HttpRequest.fromMatchable(request);
		const matched = this._routes.match(req);
		const route = matched.cloneForRequest().bind(req);
		const context = new HttpContext(req, new HttpResponseFactory(), route, this._container);
		const middlewares = this.gatherRouteMiddleware(route);

		const runner = async (ctx: unknown): Promise<unknown> => {
			return route.run(ctx);
		};

		const pipeline = this.buildPipeline(middlewares, runner);

		const value = await pipeline(context);

		return { route, value, context };
	}

	gatherRouteMiddleware(route: Route): unknown[] {
		const raw = route.gatherMiddleware();
		const resolved: unknown[] = [];

		for (const m of raw) {
			if (typeof m === 'string') {
				if (this._middlewareGroups.has(m)) {
					resolved.push(...this._middlewareGroups.get(m)!);
				} else if (this._middleware.has(m)) {
					resolved.push(this._middleware.get(m)!);
				} else {
					resolved.push(m);
				}
			} else {
				resolved.push(m);
			}
		}

		return resolved;
	}

	private buildPipeline(middlewares: unknown[], destination: (req: unknown) => unknown): (req: unknown) => Promise<unknown> {
		let current = destination;

		for (let i = middlewares.length - 1; i >= 0; i--) {
			const mw = middlewares[i];
			const next = current;

			current = async (req: unknown) => {
				if (typeof mw === 'function') {
					return mw(req, next);
				}

				if (typeof mw === 'object' && mw !== null && 'handle' in mw && typeof (mw as MiddlewareObject).handle === 'function') {
					return (mw as MiddlewareObject).handle(req, next);
				}

				return next(req);
			};
		}

		return async (req: unknown) => current(req);
	}

	private createRoute(methods: string | readonly string[], uri: string, action?: RouteAction): Route {
		const route = new Route(methods, uri, action);

		if (this.hasGroupStack()) {
			this.mergeGroupAttributesIntoRoute(route);
		}

		this.addWhereClausesToRoute(route);

		return route;
	}

	private hasGroupStack(): boolean {
		return this._groupStack.length > 0;
	}

	private updateGroupStack(attributes: Record<string, unknown>): void {
		if (this.hasGroupStack()) {
			const parent = this._groupStack[this._groupStack.length - 1];

			this._groupStack.push(mergeRouteGroup(attributes, parent));
		} else {
			this._groupStack.push(mergeRouteGroup(attributes, {}));
		}
	}

	private mergeGroupAttributesIntoRoute(route: Route): void {
		const group = this._groupStack[this._groupStack.length - 1];

		if (group.prefix && typeof group.prefix === 'string') {
			route.prefix(group.prefix);
		}

		if (group.domain && typeof group.domain === 'string') {
			route.domain(group.domain);
		}

		if (group.as && typeof group.as === 'string') {
			route.name(group.as);
		}

		if (group.middleware) {
			route.middleware(group.middleware);
		}

		if (group.where && typeof group.where === 'object') {
			route.where(group.where as Record<string, string>);
		}
	}

	private addWhereClausesToRoute(route: Route): void {
		for (const [k, v] of Object.entries(this._patterns)) {
			route.where(k, v);
		}
	}
}

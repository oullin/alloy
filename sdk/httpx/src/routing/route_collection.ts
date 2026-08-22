import { HTTP_VERBS, MethodNotAllowedError, RouteNotFoundError } from '#httpx/errors';
import type { MatchableRequest } from '#httpx/routing/matching/contracts';
import { Route } from '#httpx/routing/route';

export class RouteCollection {
	private routesByMethod: Map<string, Route[]> = new Map();
	private routesByMethodKey: Map<string, Map<string, number>> = new Map();
	private allRoutes: Route[] = [];
	private allKeys: Map<string, number> = new Map();
	private nameList: Map<string, Route> = new Map();
	private actionList: Map<string, Route> = new Map();

	add(route: Route): Route {
		this.addToCollections(route);
		this.addLookups(route);

		return route;
	}

	private addToCollections(route: Route): void {
		const domainAndUri = route.getDomain() + route.uri;

		for (const method of route.methods()) {
			const key = domainAndUri;

			let methodRoutes = this.routesByMethod.get(method);
			let methodKeys = this.routesByMethodKey.get(method);

			if (!methodRoutes || !methodKeys) {
				methodRoutes = [];
				methodKeys = new Map();
				this.routesByMethod.set(method, methodRoutes);
				this.routesByMethodKey.set(method, methodKeys);
			}

			if (methodKeys.has(key)) {
				const idx = methodKeys.get(key)!;

				methodRoutes[idx] = route;
			} else {
				const idx = methodRoutes.length;

				methodRoutes.push(route);
				methodKeys.set(key, idx);
			}
		}

		const allKey = route.methods().join('|') + domainAndUri;

		if (this.allKeys.has(allKey)) {
			const idx = this.allKeys.get(allKey)!;

			this.allRoutes[idx] = route;

			return;
		}

		this.allKeys.set(allKey, this.allRoutes.length);
		this.allRoutes.push(route);
	}

	private addLookups(route: Route): void {
		const name = route.getName();

		if (name !== '' && !this.nameList.has(name)) {
			this.nameList.set(name, route);
		}

		const handler = route.action.handler;

		if (typeof handler === 'string' && handler !== '') {
			const clean = handler.replace(/^\\+/u, '');

			if (!this.actionList.has(clean)) {
				this.actionList.set(clean, route);
			}
		}
	}

	refreshNameLookups(): void {
		this.nameList.clear();

		for (const r of this.allRoutes) {
			const name = r.getName();

			if (name !== '' && !this.nameList.has(name)) {
				this.nameList.set(name, r);
			}
		}
	}

	refreshActionLookups(): void {
		this.actionList.clear();

		for (const r of this.allRoutes) {
			const handler = r.action.handler;

			if (typeof handler === 'string' && handler !== '') {
				const clean = handler.replace(/^\\+/u, '');

				if (!this.actionList.has(clean)) {
					this.actionList.set(clean, r);
				}
			}
		}
	}

	get(method?: string): Route[] {
		if (!method || method === '') {
			return [...this.allRoutes];
		}

		return [...(this.routesByMethod.get(method.toUpperCase()) ?? [])];
	}

	getRoutes(): Route[] {
		return [...this.allRoutes];
	}

	all(): Route[] {
		return [...this.allRoutes];
	}

	getByName(name: string): Route | undefined {
		if (this.nameList.has(name)) {
			return this.nameList.get(name);
		}

		this.refreshNameLookups();

		return this.nameList.get(name);
	}

	hasNamedRoute(name: string): boolean {
		return this.getByName(name) !== undefined;
	}

	getByAction(action: string): Route | undefined {
		if (this.actionList.has(action)) {
			return this.actionList.get(action);
		}

		this.refreshActionLookups();

		return this.actionList.get(action);
	}

	count(): number {
		return this.allRoutes.length;
	}

	match(request: MatchableRequest): Route {
		const routes = this.get(request.method());
		const matched = this.matchAgainstRoutes(routes, request, true);

		return this.handleMatchedRoute(request, matched);
	}

	private matchAgainstRoutes(routes: Route[], request: MatchableRequest, includingMethod: boolean): Route | null {
		let fallback: Route | null = null;

		for (const route of routes) {
			if (route.matches(request, includingMethod)) {
				const candidate = route.cloneForRequest();

				if (route.isFallback) {
					if (fallback === null) {
						fallback = candidate;
					}

					continue;
				}

				return candidate;
			}
		}

		return fallback;
	}

	private handleMatchedRoute(request: MatchableRequest, matched: Route | null): Route {
		if (matched !== null) {
			return matched.bind(request);
		}

		const others = this.checkForAlternateVerbs(request);

		if (others.length > 0) {
			return this.getRouteForMethods(request, others);
		}

		throw new RouteNotFoundError(request.path());
	}

	private checkForAlternateVerbs(request: MatchableRequest): string[] {
		const current = request.method().toUpperCase();
		const others: string[] = [];

		for (const m of HTTP_VERBS) {
			if (m === current) {
				continue;
			}

			const matched = this.matchAgainstRoutes(this.get(m), request, false);

			if (matched !== null) {
				others.push(m);
			}
		}

		return others;
	}

	private getRouteForMethods(request: MatchableRequest, methods: string[]): Route {
		if (request.method().toUpperCase() === 'OPTIONS') {
			const optionsRoute = new Route(['OPTIONS'], request.path(), () => {
				return new Response(null, {
					status: 200,
					headers: {
						Allow: methods.join(', '),
					},
				});
			});

			return optionsRoute.bind(request);
		}

		throw new MethodNotAllowedError(request.path(), methods);
	}
}

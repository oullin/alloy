import { RouteCollection } from '#httpx/routing/route_collection';
import type { Route } from '#httpx/routing/route';

export interface ResourceOptions {
	only?: string[];
	except?: string[];
	names?: Record<string, string> | string;
	parameters?: Record<string, string> | string;
	as?: string;
	shallow?: boolean;
	creatable?: boolean;
	destroyable?: boolean;
	[key: string]: unknown;
}

export interface RouterLike {
	get(uri: string, action?: unknown): Route;
	post(uri: string, action?: unknown): Route;
	put(uri: string, action?: unknown): Route;
	patch(uri: string, action?: unknown): Route;
	delete(uri: string, action?: unknown): Route;
	match(methods: string[], uri: string, action?: unknown): Route;
	group(attributes: Record<string, unknown>, callback: (router: RouterLike) => void): unknown;
}

const RESOURCE_DEFAULTS = ['index', 'create', 'store', 'show', 'edit', 'update', 'destroy'];
const SINGLETON_DEFAULTS = ['show', 'edit', 'update'];

export class ResourceRegistrar {
	constructor(private readonly router: RouterLike) {}

	register(name: string, handler: unknown, options: ResourceOptions = {}): RouteCollection {
		if (name.includes('/')) {
			return this.prefixedResource(name, handler, options);
		}

		const base = this.getResourceWildcard(this.lastSegment(name, '.'), options);
		const methods = this.getResourceMethods(RESOURCE_DEFAULTS, options);
		const collection = new RouteCollection();

		for (const m of methods) {
			let route: Route | null = null;

			switch (m) {
				case 'index':
					route = this.addResourceIndex(name, base, handler, options);
					break;

				case 'create':
					route = this.addResourceCreate(name, base, handler, options);
					break;

				case 'store':
					route = this.addResourceStore(name, base, handler, options);
					break;

				case 'show':
					route = this.addResourceShow(name, base, handler, options);
					break;

				case 'edit':
					route = this.addResourceEdit(name, base, handler, options);
					break;

				case 'update':
					route = this.addResourceUpdate(name, base, handler, options);
					break;

				case 'destroy':
					route = this.addResourceDestroy(name, base, handler, options);
					break;
			}

			if (route !== null) {
				collection.add(route);
			}
		}

		return collection;
	}

	singleton(name: string, handler: unknown, options: ResourceOptions = {}): RouteCollection {
		const defaults = [...SINGLETON_DEFAULTS];

		if (options.creatable) {
			defaults.push('create', 'store', 'destroy');
		} else if (options.destroyable) {
			defaults.push('destroy');
		}

		const methods = this.getResourceMethods(defaults, options);
		const collection = new RouteCollection();
		const base = this.getResourceWildcard(this.lastSegment(name, '.'), options);

		for (const m of methods) {
			let route: Route | null = null;

			switch (m) {
				case 'create':
					route = this.addResourceCreate(name, base, handler, options);
					break;

				case 'store':
					route = this.addResourceStore(name, base, handler, options);
					break;

				case 'show':
					route = this.addSingletonShow(name, handler, options);
					break;

				case 'edit':
					route = this.addSingletonEdit(name, handler, options);
					break;

				case 'update':
					route = this.addSingletonUpdate(name, handler, options);
					break;

				case 'destroy':
					route = this.addSingletonDestroy(name, handler, options);
					break;
			}

			if (route !== null) {
				collection.add(route);
			}
		}

		return collection;
	}

	private prefixedResource(name: string, handler: unknown, options: ResourceOptions): RouteCollection {
		const segments = name.split('/');
		const resource = segments.pop()!;
		const prefix = segments.join('/');

		let collection: RouteCollection = new RouteCollection();

		this.router.group({ prefix, as: `${prefix.replace(/\//g, '.')}.` }, (nested) => {
			const registrar = new ResourceRegistrar(nested);

			collection = registrar.register(resource, handler, options);
		});

		return collection;
	}

	private addResourceIndex(name: string, _base: string, handler: unknown, options: ResourceOptions): Route {
		const uri = this.getResourceUri(name);
		const action = this.getResourceAction(name, handler, 'index', options);

		return this.router.get(uri, action);
	}

	private addResourceCreate(name: string, _base: string, handler: unknown, options: ResourceOptions): Route {
		const uri = `${this.getResourceUri(name)}/create`;
		const action = this.getResourceAction(name, handler, 'create', options);

		return this.router.get(uri, action);
	}

	private addResourceStore(name: string, _base: string, handler: unknown, options: ResourceOptions): Route {
		const uri = this.getResourceUri(name);
		const action = this.getResourceAction(name, handler, 'store', options);

		return this.router.post(uri, action);
	}

	private addResourceShow(name: string, base: string, handler: unknown, options: ResourceOptions): Route {
		const uri = `${this.getResourceUri(name)}/{${base}}`;
		const action = this.getResourceAction(name, handler, 'show', options);

		return this.router.get(uri, action);
	}

	private addResourceEdit(name: string, base: string, handler: unknown, options: ResourceOptions): Route {
		const uri = `${this.getResourceUri(name)}/{${base}}/edit`;
		const action = this.getResourceAction(name, handler, 'edit', options);

		return this.router.get(uri, action);
	}

	private addResourceUpdate(name: string, base: string, handler: unknown, options: ResourceOptions): Route {
		const uri = `${this.getResourceUri(name)}/{${base}}`;
		const action = this.getResourceAction(name, handler, 'update', options);

		return this.router.match(['PUT', 'PATCH'], uri, action);
	}

	private addResourceDestroy(name: string, base: string, handler: unknown, options: ResourceOptions): Route {
		const uri = `${this.getResourceUri(name)}/{${base}}`;
		const action = this.getResourceAction(name, handler, 'destroy', options);

		return this.router.delete(uri, action);
	}

	private addSingletonShow(name: string, handler: unknown, options: ResourceOptions): Route {
		const uri = this.getResourceUri(name);
		const action = this.getResourceAction(name, handler, 'show', options);

		return this.router.get(uri, action);
	}

	private addSingletonEdit(name: string, handler: unknown, options: ResourceOptions): Route {
		const uri = `${this.getResourceUri(name)}/edit`;
		const action = this.getResourceAction(name, handler, 'edit', options);

		return this.router.get(uri, action);
	}

	private addSingletonUpdate(name: string, handler: unknown, options: ResourceOptions): Route {
		const uri = this.getResourceUri(name);
		const action = this.getResourceAction(name, handler, 'update', options);

		return this.router.match(['PUT', 'PATCH'], uri, action);
	}

	private addSingletonDestroy(name: string, handler: unknown, options: ResourceOptions): Route {
		const uri = this.getResourceUri(name);
		const action = this.getResourceAction(name, handler, 'destroy', options);

		return this.router.delete(uri, action);
	}

	private getResourceUri(name: string): string {
		if (!name.includes('.')) {
			return name;
		}

		const segments = name.split('.');

		let uri = '';

		for (let i = 0; i < segments.length; i++) {
			const s = segments[i];

			if (i === segments.length - 1) {
				uri += `/${s}`;
			} else {
				uri += `/${s}/{${this.getResourceWildcard(s)}}`;
			}
		}

		return uri.replace(/^\/+/u, '');
	}

	private getResourceWildcard(value: string, options: ResourceOptions = {}): string {
		if (typeof options.parameters === 'object' && options.parameters !== null && value in options.parameters) {
			return options.parameters[value];
		}

		if (typeof options.parameters === 'string' && options.parameters !== '') {
			return options.parameters;
		}

		const trimmed = value.endsWith('s') ? value.slice(0, -1) : value;

		return trimmed.replace(/-/g, '_');
	}

	private getResourceMethods(defaults: string[], options: ResourceOptions): string[] {
		let methods = [...defaults];

		if (options.only && options.only.length > 0) {
			methods = methods.filter((m) => options.only!.includes(m));
		}

		if (options.except && options.except.length > 0) {
			methods = methods.filter((m) => !options.except!.includes(m));
		}

		return methods;
	}

	private getResourceAction(name: string, handler: unknown, action: string, options: ResourceOptions): Record<string, unknown> {
		const routeName = this.getResourceRouteName(name, action, options);

		if (typeof handler === 'object' && handler !== null) {
			const candidate = (handler as Record<string, unknown>)[action];

			return { uses: candidate ?? handler, as: routeName };
		}

		if (typeof handler === 'string') {
			return { uses: `${handler}@${action}`, as: routeName };
		}

		return { uses: handler, as: routeName };
	}

	private getResourceRouteName(name: string, action: string, options: ResourceOptions): string {
		if (typeof options.names === 'object' && options.names !== null && action in options.names) {
			return options.names[action];
		}

		if (typeof options.names === 'string' && options.names !== '') {
			return `${options.names}.${action}`;
		}

		const prefix = options.as ? `${options.as}.` : '';

		return `${prefix}${name}.${action}`;
	}

	private lastSegment(str: string, separator: string): string {
		const parts = str.split(separator);

		return parts[parts.length - 1];
	}
}

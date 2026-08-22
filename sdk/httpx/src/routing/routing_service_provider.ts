import { createServiceToken, type Application, type ServiceProvider, type ServiceToken } from '@hara/sdk-container';
import { Router } from './router.js';
import { UrlGenerator } from './url_generator.js';

export const routerToken: ServiceToken<Router> = createServiceToken<Router>('router');

export const urlGeneratorToken: ServiceToken<UrlGenerator> = createServiceToken<UrlGenerator>('url_generator');

export class RoutingServiceProvider implements ServiceProvider {
	readonly provides: readonly [ServiceToken<Router>, ServiceToken<UrlGenerator>] = [routerToken, urlGeneratorToken];

	register(app: Application): void {
		app.singleton(routerToken, (c) => new Router(c));
		app.singleton(urlGeneratorToken, (c) => {
			const router = c.make(routerToken);

			return new UrlGenerator(router.getRoutes());
		});
	}

	boot(_app: Application): void {
		// No-op for v1 routing provider
	}
}

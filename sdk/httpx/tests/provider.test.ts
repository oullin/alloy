import { describe, expect, it } from 'vite-plus/test';
import { Application } from '@hara/sdk-container';

import { Router, RoutingServiceProvider, UrlGenerator, routerToken, urlGeneratorToken } from '#httpx/routing/index.js';

describe('RoutingServiceProvider', () => {
	it('binds router and url generator into Application', () => {
		const app = new Application();

		app.register(new RoutingServiceProvider());
		app.boot();

		const router = app.make(routerToken);

		expect(router).toBeInstanceOf(Router);

		const urlGenerator = app.make(urlGeneratorToken);

		expect(urlGenerator).toBeInstanceOf(UrlGenerator);

		// Assert idempotency
		app.boot();
		expect(app.make(routerToken)).toBe(router);
		expect(app.make(urlGeneratorToken)).toBe(urlGenerator);
	});
});

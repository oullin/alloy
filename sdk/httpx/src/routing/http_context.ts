import type { Container } from '@hara/sdk-container';
import type { HttpRequest } from '#httpx/foundation/http_request.js';
import type { HttpResponseFactory } from '#httpx/foundation/http_response_factory.js';
import type { Route } from './route.js';

export class HttpContext {
	constructor(
		readonly request: HttpRequest,
		readonly response: HttpResponseFactory,
		readonly route: Route | undefined,
		readonly container?: Container,
	) {}
}

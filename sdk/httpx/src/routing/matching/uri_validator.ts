import type { MatchableRequest, MatchableRoute, RouteMatchingValidator } from '#httpx/routing/matching/contracts';

export class UriValidator implements RouteMatchingValidator {
	matches(route: MatchableRoute, request: MatchableRequest): boolean {
		const c = route.compiled();

		if (c === null) {
			return false;
		}

		let path = request.pathInfo().replace(/\/+$/u, '');

		if (path === '') {
			path = '/';
		}

		try {
			path = decodeURIComponent(path);
		} catch {
			// keep path as-is if malformed percent encoding
		}

		return c.compiledRegex.test(path);
	}
}

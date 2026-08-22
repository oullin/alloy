import type { MatchableRequest, MatchableRoute, RouteMatchingValidator } from '#httpx/routing/matching/contracts';

export class HostValidator implements RouteMatchingValidator {
	matches(route: MatchableRoute, request: MatchableRequest): boolean {
		const c = route.compiled();

		if (c === null || c.hostRegex === '' || c.compiledHostRegex === null) {
			return true;
		}

		return c.compiledHostRegex.test(request.host());
	}
}

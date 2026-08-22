import type { MatchableRequest, MatchableRoute, RouteMatchingValidator } from '#httpx/routing/matching/contracts';

export class SchemeValidator implements RouteMatchingValidator {
	matches(route: MatchableRoute, request: MatchableRequest): boolean {
		if (route.httpOnly()) {
			return !request.secure();
		}

		if (route.secure()) {
			return request.secure();
		}

		return true;
	}
}

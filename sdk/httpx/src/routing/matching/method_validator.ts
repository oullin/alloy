import type { MatchableRequest, MatchableRoute, RouteMatchingValidator } from '#httpx/routing/matching/contracts';

export class MethodValidator implements RouteMatchingValidator {
	matches(route: MatchableRoute, request: MatchableRequest): boolean {
		const rm = request.method().toUpperCase();

		return route.methods().includes(rm);
	}
}

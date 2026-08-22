import { HttpResponse } from '#httpx/foundation/http_response.js';

export type NextFunction = (req?: unknown) => unknown;

export class RecoveryMiddleware {
	constructor(private readonly _report?: (error: unknown) => void) {}

	async handle(request: unknown, next: NextFunction): Promise<unknown> {
		try {
			return await next(request);
		} catch (error) {
			this._report?.(error);

			return HttpResponse.fromError(error);
		}
	}
}

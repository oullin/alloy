import { HttpResponse } from '#httpx/foundation/http_response.js';
import type { NextFunction } from './recovery.js';

export type FrameGuardMode = 'DENY' | 'SAMEORIGIN';

export class FrameGuardMiddleware {
	constructor(private readonly _mode: FrameGuardMode = 'SAMEORIGIN') {}

	async handle(requestOrContext: unknown, next: NextFunction): Promise<unknown> {
		const response = await next(requestOrContext);

		if (response instanceof HttpResponse) {
			response.header('X-Frame-Options', this._mode);
		}

		return response;
	}
}

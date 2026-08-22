import { HttpResponseError } from '#httpx/errors.js';
import { HttpRequest } from '#httpx/foundation/http_request.js';
import type { NextFunction } from './recovery.js';

export class ValidatePathEncodingMiddleware {
	async handle(requestOrContext: unknown, next: NextFunction): Promise<unknown> {
		const req = this.extractRequest(requestOrContext);

		if (req) {
			const path = req.pathInfo();
			// Check for malformed percent encoding
			try {
				decodeURIComponent(path);
			} catch {
				throw new HttpResponseError(400, 'Bad Request');
			}
		}

		return next(requestOrContext);
	}

	private extractRequest(val: unknown): HttpRequest | null {
		if (val instanceof HttpRequest) {
			return val;
		}

		if (typeof val === 'object' && val !== null && 'request' in val) {
			const inner = (val as { request: unknown }).request;

			if (inner instanceof HttpRequest) {
				return inner;
			}
		}

		return null;
	}
}

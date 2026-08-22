import { HttpResponseError } from '#httpx/errors.js';
import { HttpRequest } from '#httpx/foundation/http_request.js';
import type { NextFunction } from './recovery.js';

export class ValidatePostSizeMiddleware {
	constructor(private readonly _maxBytes: number) {}

	async handle(requestOrContext: unknown, next: NextFunction): Promise<unknown> {
		const req = this.extractRequest(requestOrContext);

		if (req) {
			const cl = req.header('content-length');

			if (cl) {
				const size = parseInt(cl, 10);

				if (!Number.isNaN(size) && size > this._maxBytes) {
					throw new HttpResponseError(413, 'Post body too large');
				}
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

import { HttpResponse } from '#httpx/foundation/http_response.js';
import type { NextFunction } from './recovery.js';

export interface CacheOptions {
	public?: boolean;
	private?: boolean;
	noCache?: boolean;
	noStore?: boolean;
	maxAge?: number;
	sMaxAge?: number;
	mustRevalidate?: boolean;
}

export class CacheHeadersMiddleware {
	constructor(private readonly _opts: CacheOptions = {}) {}

	async handle(requestOrContext: unknown, next: NextFunction): Promise<unknown> {
		const response = await next(requestOrContext);

		if (response instanceof HttpResponse) {
			const directives: string[] = [];

			if (this._opts.public) {
				directives.push('public');
			}

			if (this._opts.private) {
				directives.push('private');
			}

			if (this._opts.noCache) {
				directives.push('no-cache');
			}

			if (this._opts.noStore) {
				directives.push('no-store');
			}

			if (this._opts.maxAge !== undefined && this._opts.maxAge >= 0) {
				directives.push(`max-age=${this._opts.maxAge}`);
			}

			if (this._opts.sMaxAge !== undefined && this._opts.sMaxAge >= 0) {
				directives.push(`s-maxage=${this._opts.sMaxAge}`);
			}

			if (this._opts.mustRevalidate) {
				directives.push('must-revalidate');
			}

			if (directives.length > 0) {
				response.header('Cache-Control', directives.join(', '));
			}
		}

		return response;
	}
}

import { HttpRequest } from '#httpx/foundation/http_request.js';
import { HttpResponse } from '#httpx/foundation/http_response.js';
import type { NextFunction } from './recovery.js';

export interface CorsOptions {
	allowedOrigins?: string[];
	allowedMethods?: string[];
	allowedHeaders?: string[];
	exposedHeaders?: string[];
	maxAge?: number;
	allowCredentials?: boolean;
}

export class CorsMiddleware {
	constructor(private readonly _opts: CorsOptions = {}) {}

	async handle(requestOrContext: unknown, next: NextFunction): Promise<unknown> {
		const req = this.extractRequest(requestOrContext);

		if (!req) {
			return next(requestOrContext);
		}

		const origin = req.header('origin');

		if (!origin || !this.isAllowedOrigin(origin)) {
			return next(requestOrContext);
		}

		// Handle OPTIONS preflight
		if (req.isMethod('OPTIONS') && req.hasHeader('access-control-request-method')) {
			const preflight = new HttpResponse(null, 204);

			this.applyCorsHeaders(preflight, origin);
			this.applyPreflightHeaders(preflight, req);

			return preflight;
		}

		const response = await next(requestOrContext);

		if (response instanceof HttpResponse) {
			this.applyCorsHeaders(response, origin);
		}

		return response;
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

	private isAllowedOrigin(origin: string): boolean {
		const allowed = this._opts.allowedOrigins ?? ['*'];

		if (allowed.includes('*')) {
			return true;
		}

		return allowed.includes(origin);
	}

	private applyCorsHeaders(resp: HttpResponse, origin: string): void {
		const allowed = this._opts.allowedOrigins ?? ['*'];

		if (allowed.includes('*')) {
			resp.header('Access-Control-Allow-Origin', '*');
		} else {
			resp.header('Access-Control-Allow-Origin', origin);
			resp.header('Vary', 'Origin');
		}

		if (this._opts.exposedHeaders && this._opts.exposedHeaders.length > 0) {
			resp.header('Access-Control-Expose-Headers', this._opts.exposedHeaders.join(', '));
		}

		if (this._opts.allowCredentials) {
			resp.header('Access-Control-Allow-Credentials', 'true');
		}
	}

	private applyPreflightHeaders(resp: HttpResponse, req: HttpRequest): void {
		if (this._opts.allowedMethods && this._opts.allowedMethods.length > 0) {
			resp.header('Access-Control-Allow-Methods', this._opts.allowedMethods.join(', '));
		}

		const requestedHeaders = req.header('access-control-request-headers');

		if (requestedHeaders) {
			if (this._opts.allowedHeaders?.includes('*') || !this._opts.allowedHeaders) {
				resp.header('Access-Control-Allow-Headers', requestedHeaders);
			} else {
				resp.header('Access-Control-Allow-Headers', this._opts.allowedHeaders.join(', '));
			}
		}

		if (this._opts.maxAge !== undefined && this._opts.maxAge >= 0) {
			resp.header('Access-Control-Max-Age', String(this._opts.maxAge));
		}
	}
}

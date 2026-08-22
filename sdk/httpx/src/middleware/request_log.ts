import { HttpRequest } from '#httpx/foundation/http_request.js';
import { HttpResponse } from '#httpx/foundation/http_response.js';
import type { NextFunction } from './recovery.js';

export interface Logger {
	info(message: string, context?: Record<string, unknown>): void;
	error(message: string, context?: Record<string, unknown>): void;
}

export interface RequestLogOptions {
	skipPaths?: string[];
	logger?: Logger;
}

export class RequestLogMiddleware {
	constructor(private readonly _opts: RequestLogOptions = {}) {}

	async handle(requestOrContext: unknown, next: NextFunction): Promise<unknown> {
		const req = this.extractRequest(requestOrContext);

		if (req && this.shouldSkip(req.pathInfo())) {
			return next(requestOrContext);
		}

		const start = Date.now();

		const response = await next(requestOrContext);

		const duration = Date.now() - start;

		if (req && this._opts.logger) {
			const statusCode = response instanceof HttpResponse ? response.statusCode() : 200;

			this._opts.logger.info('http: request', {
				method: req.method(),
				path: req.pathInfo(),
				status: statusCode,
				duration_ms: duration,
			});
		}

		return response;
	}

	private shouldSkip(path: string): boolean {
		return this._opts.skipPaths?.includes(path) ?? false;
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

import { HttpRequest } from '#httpx/foundation/http_request.js';
import type { NextFunction } from './recovery.js';

export const HEADER_FORWARDED_FOR = 1 << 0;

export const HEADER_FORWARDED_HOST = 1 << 1;

export const HEADER_FORWARDED_PROTO = 1 << 2;

export const HEADER_FORWARDED_PORT = 1 << 3;

export const HEADER_FORWARDED_ALL = HEADER_FORWARDED_FOR | HEADER_FORWARDED_HOST | HEADER_FORWARDED_PROTO | HEADER_FORWARDED_PORT;

export class TrustProxiesMiddleware {
	private readonly _proxies: string[];
	private readonly _headers: number;
	private readonly _trustAll: boolean;

	constructor(proxies: string[] = ['*'], headers = HEADER_FORWARDED_ALL) {
		this._proxies = proxies;
		this._headers = headers;
		this._trustAll = proxies.includes('*') || proxies.includes('**');
	}

	async handle(requestOrContext: unknown, next: NextFunction): Promise<unknown> {
		const req = this.extractRequest(requestOrContext);

		if (req && !this.isTrustedProxy(req)) {
			// Strip forwarded headers if proxy is not trusted
			if (this._headers & HEADER_FORWARDED_FOR) {
				req.headers().delete('X-Forwarded-For');
			}

			if (this._headers & HEADER_FORWARDED_HOST) {
				req.headers().delete('X-Forwarded-Host');
			}

			if (this._headers & HEADER_FORWARDED_PROTO) {
				req.headers().delete('X-Forwarded-Proto');
			}

			if (this._headers & HEADER_FORWARDED_PORT) {
				req.headers().delete('X-Forwarded-Port');
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

	private isTrustedProxy(req: HttpRequest): boolean {
		if (this._trustAll) {
			return true;
		}

		const ip = req.ip();

		return this._proxies.includes(ip);
	}
}

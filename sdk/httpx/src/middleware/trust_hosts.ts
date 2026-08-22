import { HttpResponseError } from '#httpx/errors.js';
import { HttpRequest } from '#httpx/foundation/http_request.js';
import type { NextFunction } from './recovery.js';

export class TrustHostsMiddleware {
	private readonly _hosts: string[];

	constructor(...hosts: string[]) {
		this._hosts = hosts.map((h) => h.toLowerCase());
	}

	async handle(requestOrContext: unknown, next: NextFunction): Promise<unknown> {
		const req = this.extractRequest(requestOrContext);

		if (req && !this.isTrustedHost(req.host())) {
			throw new HttpResponseError(400, 'Bad Request');
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

	private isTrustedHost(host: string): boolean {
		let cleanHost = host.toLowerCase();

		const colonIdx = cleanHost.lastIndexOf(':');

		if (colonIdx !== -1) {
			cleanHost = cleanHost.slice(0, colonIdx);
		}

		for (const allowed of this._hosts) {
			if (allowed === cleanHost) {
				return true;
			}

			if (allowed.startsWith('*.')) {
				const suffix = allowed.slice(1); // .example.com

				if (cleanHost.endsWith(suffix)) {
					return true;
				}
			}
		}

		return false;
	}
}

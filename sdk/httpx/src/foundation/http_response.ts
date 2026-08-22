import { HttpResponseError } from '#httpx/errors.js';

export interface CookieOptions {
	name: string;
	value: string;
	path?: string;
	domain?: string;
	maxAge?: number;
	expires?: Date;
	secure?: boolean;
	httpOnly?: boolean;
	sameSite?: 'strict' | 'lax' | 'none';
}

export class HttpResponse {
	private _statusCode = 200;
	private _headers: Record<string, string> = {};
	private _cookies: CookieOptions[] = [];
	private _body: BodyInit | null = null;
	private _exception: unknown = null;
	private _original: unknown = null;

	constructor(body: BodyInit | null = null, statusCode = 200, headers: Record<string, string> = {}) {
		this._body = body;
		this._statusCode = statusCode;
		this._headers = { ...headers };
	}

	status(code: number): this {
		this._statusCode = code;

		return this;
	}

	statusCode(): number {
		return this._statusCode;
	}

	header(key: string, value: string): this {
		this._headers[key] = value;

		return this;
	}

	withHeaders(headers: Record<string, string>): this {
		Object.assign(this._headers, headers);

		return this;
	}

	withoutHeader(key: string): this {
		delete this._headers[key];

		return this;
	}

	getHeader(key: string): string | undefined {
		for (const [k, v] of Object.entries(this._headers)) {
			if (k.toLowerCase() === key.toLowerCase()) {
				return v;
			}
		}

		return undefined;
	}

	getHeaders(): Readonly<Record<string, string>> {
		return this._headers;
	}

	cookie(c: CookieOptions): this {
		this._cookies.push(c);

		return this;
	}

	withoutCookie(name: string, path = '/'): this {
		this._cookies.push({
			name,
			value: '',
			path,
			maxAge: -1,
		});

		return this;
	}

	withException(err: unknown): this {
		this._exception = err;

		return this;
	}

	getException(): unknown {
		return this._exception;
	}

	setOriginal(orig: unknown): this {
		this._original = orig;

		return this;
	}

	getOriginal(): unknown {
		return this._original;
	}

	setBody(body: BodyInit | null): this {
		this._body = body;

		return this;
	}

	getBody(): BodyInit | null {
		return this._body;
	}

	toFetch(): Response {
		const headers = new Headers(this._headers);

		for (const c of this._cookies) {
			let cookieStr = `${encodeURIComponent(c.name)}=${encodeURIComponent(c.value)}`;

			if (c.path) {
				cookieStr += `; Path=${c.path}`;
			}

			if (c.domain) {
				cookieStr += `; Domain=${c.domain}`;
			}

			if (c.maxAge !== undefined) {
				cookieStr += `; Max-Age=${c.maxAge}`;
			}

			if (c.expires) {
				cookieStr += `; Expires=${c.expires.toUTCString()}`;
			}

			if (c.secure) {
				cookieStr += `; Secure`;
			}

			if (c.httpOnly) {
				cookieStr += `; HttpOnly`;
			}

			if (c.sameSite) {
				cookieStr += `; SameSite=${c.sameSite}`;
			}

			headers.append('Set-Cookie', cookieStr);
		}

		return new Response(this._body, {
			status: this._statusCode,
			headers,
		});
	}

	static fromError(err: unknown): HttpResponse {
		if (err instanceof HttpResponseError) {
			return new HttpResponse(JSON.stringify({ error: err.message }), err.statusCode, { 'Content-Type': 'application/json; charset=utf-8', ...err.headers }).withException(err);
		}

		return new HttpResponse(JSON.stringify({ error: 'internal server error' }), 500, { 'Content-Type': 'application/json; charset=utf-8' }).withException(err);
	}

	static json(data: unknown, status = 200, headers: Record<string, string> = {}): HttpResponse {
		const body = JSON.stringify(data);

		const resp = new HttpResponse(body, status, {
			'Content-Type': 'application/json; charset=utf-8',
			...headers,
		});

		resp.setOriginal(data);

		return resp;
	}

	static text(content: string, status = 200, headers: Record<string, string> = {}): HttpResponse {
		return new HttpResponse(content, status, {
			'Content-Type': 'text/plain; charset=utf-8',
			...headers,
		});
	}

	static html(content: string, status = 200, headers: Record<string, string> = {}): HttpResponse {
		return new HttpResponse(content, status, {
			'Content-Type': 'text/html; charset=utf-8',
			...headers,
		});
	}

	static noContent(status = 204, headers: Record<string, string> = {}): HttpResponse {
		return new HttpResponse(null, status, headers);
	}

	static redirect(url: string, status = 302, headers: Record<string, string> = {}): HttpResponse {
		return new HttpResponse(null, status, {
			Location: url,
			...headers,
		});
	}
}

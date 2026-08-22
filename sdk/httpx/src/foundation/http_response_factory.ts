import { HttpResponse } from './http_response.js';

export class HttpResponseFactory {
	json(data: unknown, status = 200, headers: Record<string, string> = {}): HttpResponse {
		return HttpResponse.json(data, status, headers);
	}

	text(content: string, status = 200, headers: Record<string, string> = {}): HttpResponse {
		return HttpResponse.text(content, status, headers);
	}

	html(content: string, status = 200, headers: Record<string, string> = {}): HttpResponse {
		return HttpResponse.html(content, status, headers);
	}

	noContent(status = 204, headers: Record<string, string> = {}): HttpResponse {
		return HttpResponse.noContent(status, headers);
	}

	redirect(url: string, status = 302, headers: Record<string, string> = {}): HttpResponse {
		return HttpResponse.redirect(url, status, headers);
	}

	make(body?: BodyInit | null, status = 200, headers: Record<string, string> = {}): HttpResponse {
		return new HttpResponse(body ?? null, status, headers);
	}
}

export function mergeRouteGroup(newAttrs: Record<string, unknown> = {}, oldAttrs: Record<string, unknown> = {}, prependExistingPrefix = true): Record<string, unknown> {
	const oldCopy = { ...oldAttrs };
	const newCopy = { ...newAttrs };

	if ('domain' in newCopy) {
		delete oldCopy.domain;
	}

	if ('handler' in newCopy) {
		delete oldCopy.handler;
	}

	const formattedAs = formatAs(newCopy, oldCopy);
	const formattedPrefix = formatPrefix(newCopy, oldCopy, prependExistingPrefix);
	const formattedWhere = formatWhere(newCopy, oldCopy);

	const out: Record<string, unknown> = {};

	for (const [k, v] of Object.entries(oldCopy)) {
		if (k === 'prefix' || k === 'where' || k === 'as' || k === 'namespace') {
			continue;
		}

		out[k] = v;
	}

	for (const [k, v] of Object.entries(newCopy)) {
		if (k === 'middleware') {
			if (out.middleware !== undefined) {
				out.middleware = mergeMiddleware(out.middleware, v);
				continue;
			}
		}

		out[k] = v;
	}

	if (formattedAs !== undefined) {
		out.as = formattedAs;
	}

	if (formattedPrefix !== undefined) {
		out.prefix = formattedPrefix;
	}

	if (formattedWhere !== undefined) {
		out.where = formattedWhere;
	}

	return out;
}

function formatAs(newAttrs: Record<string, unknown>, oldAttrs: Record<string, unknown>): string | undefined {
	const newAs = (newAttrs.as as string) ?? '';
	const oldAs = (oldAttrs.as as string) ?? '';

	if (newAs !== '' && oldAs !== '') {
		return `${oldAs}${newAs}`;
	}

	return newAs !== '' ? newAs : oldAs !== '' ? oldAs : undefined;
}

function formatPrefix(newAttrs: Record<string, unknown>, oldAttrs: Record<string, unknown>, prependExisting: boolean): string | undefined {
	const oldPrefix = ((oldAttrs.prefix as string) ?? '').replace(/^\/+/u, '').replace(/\/+$/u, '');
	const newPrefix = ((newAttrs.prefix as string) ?? '').replace(/^\/+/u, '').replace(/\/+$/u, '');

	if (prependExisting) {
		if (newPrefix !== '' && oldPrefix !== '') {
			return `${oldPrefix}/${newPrefix}`;
		}

		return newPrefix !== '' ? newPrefix : oldPrefix !== '' ? oldPrefix : undefined;
	}

	if (newPrefix !== '' && oldPrefix !== '') {
		return `${newPrefix}/${oldPrefix}`;
	}

	return newPrefix !== '' ? newPrefix : oldPrefix !== '' ? oldPrefix : undefined;
}

function formatWhere(newAttrs: Record<string, unknown>, oldAttrs: Record<string, unknown>): Record<string, string> | undefined {
	const oldWhere = (oldAttrs.where as Record<string, string>) ?? {};
	const newWhere = (newAttrs.where as Record<string, string>) ?? {};

	const merged = { ...oldWhere, ...newWhere };

	return Object.keys(merged).length > 0 ? merged : undefined;
}

function mergeMiddleware(existing: unknown, adding: unknown): unknown[] {
	const exList = Array.isArray(existing) ? existing : [existing];
	const addList = Array.isArray(adding) ? adding : [adding];

	return [...exList, ...addList];
}

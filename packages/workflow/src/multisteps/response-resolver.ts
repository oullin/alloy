import { UnresolvedResponseError } from '#workflow/multisteps/errors';

export const resolveField = (raw: unknown, field: string, job = ''): unknown => {
	if (raw === null || raw === undefined) {
		throw new UnresolvedResponseError(field, job);
	}

	if (typeof raw === 'object' || typeof raw === 'function') {
		const record = raw as Record<string, unknown>;

		if (Object.hasOwn(record, field)) {
			return record[field];
		}
	}

	throw new UnresolvedResponseError(field, job);
};

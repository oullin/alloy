import type { TransitionDeclaration } from './types.js';

export const isRecord = (value: unknown): value is Record<string, unknown> => typeof value === 'object' && value !== null && !Array.isArray(value);

export const coerceTransitions = (raw: unknown): TransitionDeclaration[] => {
	if (raw === undefined || raw === null) {
		return [];
	}

	if (!Array.isArray(raw)) {
		throw new Error(`config: transitions must be a list, got ${typeof raw}`);
	}

	return raw.map((item, index) => {
		const entry = coerceStringMap(item);

		if (Object.keys(entry).length === 0 && !isRecord(item)) {
			throw new Error(`config: transition[${index}] must be an object`);
		}

		return {
			name: typeof entry.name === 'string' ? entry.name : '',
			from: coerceStringSlice(entry.from),
			to: coerceStringSlice(entry.to),
		};
	});
};

export const coerceStringSlice = (raw: unknown): string[] => {
	if (raw === undefined || raw === null) {
		return [];
	}

	if (typeof raw === 'string') {
		return [raw];
	}

	if (!Array.isArray(raw)) {
		return [];
	}

	return raw.filter((value): value is string => typeof value === 'string');
};

export const coerceStringMap = (raw: unknown): Record<string, unknown> => {
	if (!isRecord(raw)) {
		return {};
	}

	return raw;
};

import { z } from 'zod';
import type { KeyValue } from '#console/key/types';

const keyValueSchema = z.union([z.string().transform((value) => [value]), z.array(z.string()).transform((values) => [...values])]);

const keyValuesSchema = z.array(keyValueSchema).transform((values) => values.flat());

export const parseKeyMatchValues = (values: KeyValue[]): string[] => {
	const parsed = keyValuesSchema.safeParse(values);

	if (!parsed.success) {
		throw new TypeError('Key match values must be strings or string arrays.');
	}

	return parsed.data;
};

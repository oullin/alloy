import { z } from 'zod';

const multiSearchDefaultSchema = <T>(): z.ZodType<T[]> => z.array(z.unknown()) as z.ZodType<T[]>;

export const parseMultiSearchDefault = <T>(value: unknown): T[] => {
	const parsed = multiSearchDefaultSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Default selected values must be an array.');
	}

	return parsed.data;
};

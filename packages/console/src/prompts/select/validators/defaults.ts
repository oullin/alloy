import { z } from 'zod';

const multiSelectDefaultSchema = <T>(): z.ZodType<T[]> => z.array(z.unknown()) as z.ZodType<T[]>;

export const parseMultiSelectDefault = <T>(value: unknown): T[] => {
	const parsed = multiSelectDefaultSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Default selected values must be an array.');
	}

	return parsed.data;
};

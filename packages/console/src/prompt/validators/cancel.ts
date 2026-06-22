import { z } from 'zod';

const cancelValueSchema = <T>(): z.ZodType<T> => z.unknown() as z.ZodType<T>;

export const parseCancelValue = <T>(value: unknown): T => {
	const parsed = cancelValueSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Cancelled prompt values must resolve to a typed value.');
	}

	return parsed.data;
};

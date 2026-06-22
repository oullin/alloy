import { z } from 'zod';

const invalidPromptValueSchema = <T>(): z.ZodType<T | undefined> => z.unknown() as z.ZodType<T | undefined>;

export const parseInvalidPromptValue = <T>(value: unknown): T | undefined => {
	const parsed = invalidPromptValueSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Invalid prompt values must resolve to a typed value or undefined.');
	}

	return parsed.data;
};

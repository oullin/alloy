import { z } from 'zod';

const definedPromptDefaultSchema = z.unknown().refine((value) => value !== undefined);
const promptDefaultSchema = z.object({ default: definedPromptDefaultSchema }).passthrough();
const promptDefaultValueSchema = <T>(): z.ZodType<T> => z.unknown() as z.ZodType<T>;

export const hasPromptDefault = (value: unknown): boolean => {
	return promptDefaultSchema.safeParse(value).success;
};

export const parsePromptDefault = <T>(value: unknown): T => {
	const parsed = promptDefaultValueSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Prompt defaults must resolve to a typed value.');
	}

	return parsed.data;
};

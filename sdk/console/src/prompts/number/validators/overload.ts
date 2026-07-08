import { z } from 'zod';

const numberPromptLabelSchema = z.string();
const numberDefaultSchema = z.union([z.number(), z.string()]);

export const isNumberPromptLabel = (value: unknown): value is string => {
	return numberPromptLabelSchema.safeParse(value).success;
};

export const hasNumberDefaultArgument = (hasDefaultArgument: boolean, defaultValue: number | string | undefined): boolean => {
	return hasDefaultArgument && defaultValue !== undefined;
};

export const parseNumberDefault = (value: unknown, fallback: number | string): number | string => {
	const parsed = numberDefaultSchema.safeParse(value);

	return parsed.success ? parsed.data : fallback;
};

import { z } from 'zod';

const textPromptLabelSchema = z.string();

export const isTextPromptLabel = (value: unknown): value is string => {
	return textPromptLabelSchema.safeParse(value).success;
};

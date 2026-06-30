import { z } from 'zod';

const suggestPromptLabelSchema = z.string();

export const isSuggestPromptLabel = (value: unknown): value is string => {
	return suggestPromptLabelSchema.safeParse(value).success;
};

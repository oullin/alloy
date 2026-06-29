import { z } from 'zod';

const searchPromptLabelSchema = z.string();

export const isSearchPromptLabel = (value: unknown): value is string => {
	return searchPromptLabelSchema.safeParse(value).success;
};

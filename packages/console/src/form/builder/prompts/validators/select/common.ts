import { z } from 'zod';
import { choiceOptionsSchema } from '#console/concerns/validators/choice';
import type { ChoiceOptions } from '#console/types';

const selectPromptLabelSchema = z.string();
const selectPromptDefaultSchema = z.boolean();

export const isSelectPromptLabel = (value: unknown): value is string => {
	return selectPromptLabelSchema.safeParse(value).success;
};

export const isSelectPromptOptions = <TOptions>(value: TOptions | string): value is TOptions => {
	return !isSelectPromptLabel(value);
};

export const parseSelectStepName = (value: unknown): string | undefined => {
	const parsed = selectPromptLabelSchema.safeParse(value);

	return parsed.success ? parsed.data : undefined;
};

export const parseSelectChoiceOptions = <T>(value: unknown): ChoiceOptions<T> => {
	const parsed = choiceOptionsSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Choice options must be an array or record.');
	}

	return parsed.data;
};

export const parseConfirmDefault = (value: unknown, fallback: boolean): boolean => {
	const parsed = selectPromptDefaultSchema.safeParse(value);

	return parsed.success ? parsed.data : fallback;
};

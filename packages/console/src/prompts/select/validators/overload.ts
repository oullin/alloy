import { z } from 'zod';
import { choiceOptionsSchema } from '#console/concerns/validators/choice';
import type { ChoiceOptions } from '#console/types';

const selectPromptLabelSchema = z.string();
const confirmDefaultSchema = z.boolean();

export const isSelectPromptLabel = (value: unknown): value is string => {
	return selectPromptLabelSchema.safeParse(value).success;
};

export const hasSelectDefaultArgument = <T>(argumentCount: number, defaultValue: T | undefined): boolean => {
	return argumentCount >= 3 && defaultValue !== undefined;
};

export const hasConfirmDefaultArgument = (argumentCount: number, defaultValue: boolean | undefined): boolean => {
	return argumentCount >= 2 && defaultValue !== undefined;
};

export const parseSelectChoiceOptions = <T>(value: unknown): ChoiceOptions<T> => {
	const parsed = choiceOptionsSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Choice options must be an array or record.');
	}

	return parsed.data;
};

export const parseConfirmDefault = (value: unknown, fallback: boolean): boolean => {
	const parsed = confirmDefaultSchema.safeParse(value);

	return parsed.success ? parsed.data : fallback;
};

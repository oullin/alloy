import { z } from 'zod';

const basicPromptLabelSchema = z.string();

export const isBasicPromptLabel = (value: unknown): value is string => {
	return basicPromptLabelSchema.safeParse(value).success;
};

export const parseBasicStepName = (value: unknown): string | undefined => {
	const parsed = basicPromptLabelSchema.safeParse(value);

	return parsed.success ? parsed.data : undefined;
};

export const isBasicPromptOptions = <TOptions>(value: TOptions | string): value is TOptions => {
	return !isBasicPromptLabel(value);
};

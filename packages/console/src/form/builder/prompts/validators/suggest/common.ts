import { z } from 'zod';
import { functionSchema } from '#console/validators/function';
import type { MaybePromise } from '#console/types';

export type SuggestSourceCallback = (query: string) => MaybePromise<string[]>;

const suggestLabelSchema = z.string();
const suggestSourceCallbackSchema: z.ZodType<SuggestSourceCallback> = functionSchema<SuggestSourceCallback>();
const suggestSourceSchema = z.union([z.array(z.string()), suggestSourceCallbackSchema]);

export const isSuggestPromptLabel = (value: unknown): value is string => {
	return suggestLabelSchema.safeParse(value).success;
};

export const isSuggestPromptOptions = <TOptions>(value: TOptions | string): value is TOptions => {
	return !isSuggestPromptLabel(value);
};

export const parseSuggestStepName = (value: unknown): string | undefined => {
	const parsed = suggestLabelSchema.safeParse(value);

	return parsed.success ? parsed.data : undefined;
};

export const parseSuggestSource = (value: unknown): string[] | SuggestSourceCallback => {
	const parsed = suggestSourceSchema.safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Suggestion source must be an array of strings or callback.');
	}

	return parsed.data;
};

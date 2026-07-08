import { z } from 'zod';
import { choiceOptionsSchema } from '#console/concerns/validators/choice';
import { functionSchema } from '#console/validators/function';
import type { ChoiceOptions, MaybePromise } from '#console/types';

export type SearchChoiceSourceCallback<T> = (query: string) => MaybePromise<ChoiceOptions<T>>;

const searchLabelSchema = z.string();

const searchChoiceSourceCallbackSchema = <T>(): z.ZodType<SearchChoiceSourceCallback<T>> => functionSchema<SearchChoiceSourceCallback<T>>();

const searchChoiceSourceSchema = <T>(): z.ZodType<ChoiceOptions<T> | SearchChoiceSourceCallback<T>> => z.union([choiceOptionsSchema<T>(), searchChoiceSourceCallbackSchema<T>()]);

export const isSearchPromptLabel = (value: unknown): value is string => {
	return searchLabelSchema.safeParse(value).success;
};

export const isSearchPromptOptions = <TOptions>(value: TOptions | string): value is TOptions => {
	return !isSearchPromptLabel(value);
};

export const parseSearchStepName = (value: unknown): string | undefined => {
	const parsed = searchLabelSchema.safeParse(value);

	return parsed.success ? parsed.data : undefined;
};

export const parseSearchChoiceSource = <T>(value: unknown): ChoiceOptions<T> | SearchChoiceSourceCallback<T> => {
	const parsed = searchChoiceSourceSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Search choices must be an array, record, or callback.');
	}

	return parsed.data;
};

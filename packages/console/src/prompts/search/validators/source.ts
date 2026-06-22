import { z } from 'zod';
import { choiceOptionsSchema } from '#console/concerns/validators/choice';
import { functionSchema } from '#console/validators/function';
import type { ChoiceOptions, MaybePromise, SearchPromptOptions } from '#console/types';

type SearchChoiceSourceCallback<T> = (query: string) => MaybePromise<ChoiceOptions<T>>;

const searchChoiceSourceCallbackSchema = <T>(): z.ZodType<SearchChoiceSourceCallback<T>> => functionSchema<SearchChoiceSourceCallback<T>>();

const searchChoiceSourceSchema = <T>(): z.ZodType<ChoiceOptions<T> | SearchChoiceSourceCallback<T>> => z.union([choiceOptionsSchema<T>(), searchChoiceSourceCallbackSchema<T>()]);

export const isSearchChoiceSourceCallback = <T>(source: SearchPromptOptions<T>['options']): source is SearchChoiceSourceCallback<T> => {
	return searchChoiceSourceCallbackSchema<T>().safeParse(source).success;
};

export const parseSearchChoiceSource = <T>(source: unknown): ChoiceOptions<T> | SearchChoiceSourceCallback<T> => {
	const parsed = searchChoiceSourceSchema<T>().safeParse(source);

	if (!parsed.success) {
		throw new TypeError('Search choices must be an array, record, or callback.');
	}

	return parsed.data;
};

export const resolveSearchChoiceSourceOptions = async <T>(source: SearchPromptOptions<T>['options'], query: string): Promise<ChoiceOptions<T>> => {
	const parsed = parseSearchChoiceSource<T>(source);

	if (!isSearchChoiceSourceCallback(parsed)) {
		return parsed;
	}

	const options = choiceOptionsSchema<T>().safeParse(await parsed(query));

	if (!options.success) {
		throw new TypeError('Search callbacks must return choice options.');
	}

	return options.data;
};

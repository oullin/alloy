import { z } from 'zod';
import { functionSchema } from '#console/validators/function';
import type { MaybePromise } from '#console/types';

type SuggestionSourceCallback = (query: string) => MaybePromise<string[]>;

export type ResolvedSuggestionSource = {
	filter: boolean;
	options: string[];
};

const suggestionOptionsSchema = z.array(z.string());
const suggestionSourceCallbackSchema: z.ZodType<SuggestionSourceCallback> = functionSchema<SuggestionSourceCallback>();
const suggestionSourceSchema = z.union([suggestionOptionsSchema, suggestionSourceCallbackSchema]);

export const isSuggestionSourceCallback = (source: string[] | SuggestionSourceCallback): source is SuggestionSourceCallback => {
	return suggestionSourceCallbackSchema.safeParse(source).success;
};

export const parseSuggestionSource = (source: unknown): string[] | SuggestionSourceCallback => {
	const parsed = suggestionSourceSchema.safeParse(source);

	if (!parsed.success) {
		throw new TypeError('Suggestion source must be an array of strings or callback.');
	}

	return parsed.data;
};

export const resolveSuggestionSource = async (source: string[] | SuggestionSourceCallback, query: string): Promise<ResolvedSuggestionSource> => {
	const parsed = parseSuggestionSource(source);

	if (isSuggestionSourceCallback(parsed)) {
		const options = suggestionOptionsSchema.safeParse(await parsed(query));

		if (!options.success) {
			throw new TypeError('Suggestion source callbacks must return an array of strings.');
		}

		return { filter: false, options: options.data };
	}

	return { filter: true, options: parsed };
};

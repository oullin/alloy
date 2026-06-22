import { parsePromptDefault } from '#console/validators/default';
import type { MultiSearchPromptOptions, SearchPromptOptions } from '#console/types';

export type NormalizedSearchPromptOptions<T> = SearchPromptOptions<T> & {
	hasDefault: boolean;
};

export const transformSearchValue = async <T>(options: Pick<SearchPromptOptions<T>, 'transform'>, value: T): Promise<T> => {
	return options.transform ? options.transform(value) : value;
};

export const preserveSearchRetryDefault = <T>(options: NormalizedSearchPromptOptions<T>, value: T): void => {
	options.default = value;
	options.hasDefault = true;
};

export const preserveMultiSearchRetryDefault = <T>(options: MultiSearchPromptOptions<T> & { default: T[] }, value: T[]): void => {
	options.default = value;
};

export const transformedSearchDefault = async <T>(options: NormalizedSearchPromptOptions<T>): Promise<T | undefined> => {
	if (!options.hasDefault) {
		return undefined;
	}

	const rawDefault = parsePromptDefault<T>(options.default);

	try {
		return await transformSearchValue(options, rawDefault);
	} catch {
		return rawDefault;
	}
};

export const transformedMultiSearchDefault = async <T>(options: MultiSearchPromptOptions<T> & { default: T[] }): Promise<T[]> => {
	try {
		return options.transform ? await options.transform(options.default) : options.default;
	} catch {
		return options.default;
	}
};

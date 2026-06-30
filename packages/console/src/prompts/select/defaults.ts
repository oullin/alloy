import { parsePromptDefault } from '#console/validators/default';
import type { MultiSelectPromptOptions, SelectPromptOptions } from '#console/types';

export type NormalizedSelectPromptOptions<T> = SelectPromptOptions<T> & {
	hasDefault: boolean;
};

export const transformSelectValue = async <T>(options: Pick<SelectPromptOptions<T>, 'transform'>, value: T): Promise<T> => {
	return options.transform ? options.transform(value) : value;
};

export const preserveSelectRetryDefault = <T>(options: NormalizedSelectPromptOptions<T>, value: T): void => {
	options.default = value;
	options.hasDefault = true;
};

export const preserveMultiSelectRetryDefault = <T>(options: MultiSelectPromptOptions<T> & { default: T[] }, value: T[]): void => {
	options.default = value;
};

export const transformedSelectDefault = async <T>(options: NormalizedSelectPromptOptions<T>): Promise<T | undefined> => {
	if (!options.hasDefault) {
		return undefined;
	}

	const rawDefault = parsePromptDefault<T>(options.default);

	try {
		return await transformSelectValue(options, rawDefault);
	} catch {
		return rawDefault;
	}
};

export const transformedMultiSelectDefault = async <T>(options: MultiSelectPromptOptions<T> & { default: T[] }): Promise<T[]> => {
	try {
		return options.transform ? await options.transform(options.default) : options.default;
	} catch {
		return options.default;
	}
};

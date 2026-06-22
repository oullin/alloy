import { PromptValidationError } from '#console/prompt';
import { parseNumberInput } from '#console/prompts/number/validators/value';
import type { NumberPromptOptions } from '#console/types';

export type NormalizedNumberPromptOptions = NumberPromptOptions & {
	default: number | string;
	hasDefault: boolean;
};

export const transformNumberValue = async (options: Pick<NumberPromptOptions, 'transform'>, value: number | string): Promise<number | string> => {
	return options.transform ? options.transform(value) : value;
};

export const preserveNumberRetryDefault = (options: NormalizedNumberPromptOptions, value: number | string | undefined): void => {
	if (value === undefined) {
		return;
	}

	options.default = value;
	options.hasDefault = true;
};

export const numberDefault = async (options: NormalizedNumberPromptOptions): Promise<number | string> => {
	if (!options.hasDefault) {
		return '';
	}

	const result = parseNumberInput(String(options.default), options);

	if (result.error !== undefined) {
		throw new PromptValidationError(result.error);
	}

	return transformNumberValue(options, result.value ?? '');
};

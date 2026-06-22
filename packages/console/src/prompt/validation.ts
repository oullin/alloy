import { requiredMessage } from '#console/validators/required';
import { parseValidationResult } from '#console/validators/result';
import { globalValidationMessage } from '#console/prompt/validate-using';
import type { BasePromptOptions, Validator } from '#console/types';

const validationErrorMessage = (result: string | null | undefined): string | undefined => {
	if (result !== undefined && result !== null && result.length > 0) {
		return result;
	}

	return undefined;
};

export const validationMessage = async <T>(value: T, validator?: Validator<T>, options?: BasePromptOptions<T>): Promise<string | undefined> => {
	const localMessage = validationErrorMessage(parseValidationResult(validator ? await validator(value) : undefined));

	if (localMessage !== undefined) {
		return localMessage;
	}

	return validationErrorMessage(parseValidationResult(options ? await globalValidationMessage(value, options) : undefined));
};

export const ensureRequired = <T>(value: T, required?: boolean | string): string | undefined => {
	return requiredMessage(value, required);
};

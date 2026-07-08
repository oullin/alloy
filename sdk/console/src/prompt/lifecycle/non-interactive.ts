import { PromptValidationError } from '#console/prompt/error';
import { validatedPromptValue } from '#console/prompt/lifecycle/validation';
import { parsePromptDefault } from '#console/validators/default';
import type { BasePromptOptions } from '#console/types';

export const resolveNonInteractivePrompt = async <T>(options: BasePromptOptions<T>): Promise<T> => {
	const value = parsePromptDefault<T>(options.default);

	const validation = await validatedPromptValue(options, value);

	if (validation) {
		throw new PromptValidationError(validation);
	}

	return value;
};

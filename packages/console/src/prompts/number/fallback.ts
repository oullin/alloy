import { promptEnvironment } from '#console/environment';
import { PromptValidationError } from '#console/prompt';
import { renderQuestion } from '#console/theme';
import type { NumberInputOptions, NumberReadResult } from '#console/prompts/number/types';

export const readNumberFallbackValue = async (message: string, options: NumberInputOptions): Promise<NumberReadResult> => {
	const environment = promptEnvironment();

	if (!environment.input.readLine) {
		throw new PromptValidationError('The configured prompt input cannot read input.');
	}

	const answer = await environment.input.readLine(renderQuestion(message, options.hint));

	return {
		cancelled: false,
		value: answer === '' && options.hasDefault ? String(options.default) : answer,
	};
};

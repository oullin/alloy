import { promptEnvironment } from '#console/environment';
import { PromptValidationError } from '#console/prompt';
import { renderQuestion } from '#console/theme';
import type { PasswordInputOptions, PasswordReadResult } from '#console/prompts/password/types';

export const readPasswordFallbackValue = async (message: string, options: PasswordInputOptions): Promise<PasswordReadResult> => {
	const environment = promptEnvironment();

	if (!environment.input.readLine) {
		throw new PromptValidationError('The configured prompt input cannot read input.');
	}

	const answer = await environment.input.readLine(renderQuestion(message, options.hint));

	return {
		cancelled: false,
		value: answer === '' && options.default !== undefined ? options.default : answer,
	};
};

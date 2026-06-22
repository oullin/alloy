import { PromptValidationError } from '#console/prompt';
import { renderQuestion } from '#console/theme';
import type { PromptEnvironment } from '#console/contracts/environment';
import type { TypedValueReadResult } from '#console/typed-value/read';
import type { TypedValueOptions } from '#console/typed-value/types';

export const readTypedValueFallback = async (environment: PromptEnvironment, message: string, options: TypedValueOptions): Promise<TypedValueReadResult> => {
	if (!environment.input.readLine) {
		throw new PromptValidationError('The configured prompt input cannot read input.');
	}

	const answer = await environment.input.readLine(renderQuestion(message, options.hint));

	return {
		cancelled: false,
		value: answer === '' && options.default !== undefined ? options.default : answer,
	};
};

import { promptEnvironment } from '#console/environment';
import { PromptValidationError } from '#console/prompt/error';
import { renderQuestion } from '#console/theme';

export const ask = async (message: string, hint?: string): Promise<string> => {
	const environment = promptEnvironment();

	if (!environment.input.readLine) {
		throw new PromptValidationError('The configured prompt input cannot read lines.');
	}

	return environment.input.readLine(renderQuestion(message, hint));
};

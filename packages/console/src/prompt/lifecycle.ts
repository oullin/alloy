import { promptEnvironment } from '#console/environment';
import { resolveInteractivePrompt } from '#console/prompt/lifecycle/interactive';
import { resolveNonInteractivePrompt } from '#console/prompt/lifecycle/non-interactive';
import type { PromptInvalidHandler, PromptReader, PromptValidHandler } from '#console/prompt/lifecycle/types';
import type { BasePromptOptions } from '#console/types';

export const promptUntilValid = async <T>(options: BasePromptOptions<T>, read: PromptReader<T>, onValid?: PromptValidHandler<T>, onInvalid?: PromptInvalidHandler<T>): Promise<T> => {
	const environment = promptEnvironment();

	if (!environment.interactive) {
		return resolveNonInteractivePrompt(options);
	}

	return resolveInteractivePrompt(options, read, onValid, onInvalid);
};

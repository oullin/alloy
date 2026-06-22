import { promptEnvironment } from '#console/environment';
import { validatedPromptValue } from '#console/prompt/lifecycle/validation';
import { isPromptValidationError } from '#console/prompt/validators/error';
import { parseInvalidPromptValue } from '#console/prompt/validators/invalid';
import { renderError } from '#console/theme';
import type { PromptInvalidHandler, PromptReader, PromptValidHandler } from '#console/prompt/lifecycle/types';
import type { BasePromptOptions } from '#console/types';

export const resolveInteractivePrompt = async <T>(options: BasePromptOptions<T>, read: PromptReader<T>, onValid?: PromptValidHandler<T>, onInvalid?: PromptInvalidHandler<T>): Promise<T> => {
	const environment = promptEnvironment();

	let attempt = 0;

	while (true) {
		let value: T;

		try {
			value = await read(attempt);
		} catch (error) {
			if (isPromptValidationError(error)) {
				await onInvalid?.(parseInvalidPromptValue<T>(error.value));

				environment.error.write(renderError(error.message));
				attempt += 1;
				continue;
			}

			await onInvalid?.();

			throw error;
		}

		const validation = await validatedPromptValue(options, value);

		if (!validation) {
			await onValid?.(value);

			return value;
		}

		await onInvalid?.(value);

		options.default = value;
		environment.error.write(renderError(validation));
		attempt += 1;
	}
};

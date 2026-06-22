import { promptEnvironment } from '#console/environment';
import { readMultipleChoicesInteractive } from '#console/prompts/select/read-multiple/interactive';
import { readLineMultipleChoices } from '#console/prompts/select/read-multiple/line-mode';
import type { MultipleChoicesReadResult } from '#console/prompts/select/read-multiple/types';
import type { Choice, MultiSelectPromptOptions } from '#console/types';

export type { MultipleChoicesReadResult } from '#console/prompts/select/read-multiple/types';

export const readMultipleChoices = async <T>(
	message: string,
	choices: Array<Choice<T>>,
	defaults: T[] = [],
	hint?: string,
	scroll?: number,
	info?: MultiSelectPromptOptions<T>['info'],
): Promise<MultipleChoicesReadResult<T>> => {
	const environment = promptEnvironment();

	if (!environment.input.readKey) {
		return readLineMultipleChoices(message, choices, defaults, hint);
	}

	return readMultipleChoicesInteractive(async () => environment.input.readKey?.() ?? null, message, choices, defaults, scroll, info);
};

import { promptEnvironment } from '#console/environment';
import { readSelectedChoiceInteractive } from '#console/prompts/select/read-selected/interactive';
import { readSelectedChoiceLineMode } from '#console/prompts/select/read-selected/line-mode';
import type { SelectedChoiceReadResult } from '#console/prompts/select/read-selected/types';
import type { Choice, SelectPromptOptions } from '#console/types';

export type { SelectedChoiceReadResult } from '#console/prompts/select/read-selected/types';

export const readSelectedChoice = async <T>(
	message: string,
	choices: Array<Choice<T>>,
	defaultValue?: T,
	hasDefault = false,
	hint?: string,
	scroll?: number,
	info?: SelectPromptOptions<T>['info'],
): Promise<SelectedChoiceReadResult<T>> => {
	const environment = promptEnvironment();

	if (!environment.input.readKey) {
		return readSelectedChoiceLineMode(message, choices, defaultValue, hasDefault, hint);
	}

	return readSelectedChoiceInteractive(async () => environment.input.readKey?.() ?? null, message, choices, defaultValue, hasDefault, scroll, info);
};

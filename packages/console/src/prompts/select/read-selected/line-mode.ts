import { ask } from '#console/prompt';
import { renderChoices } from '#console/theme';
import { findChoice } from '#console/concerns/choices';
import { invalidSelectedChoice, selectedChoiceByDefault, selectedChoiceResult } from '#console/prompts/select/read-selected/result';
import type { SelectedChoiceReadResult } from '#console/prompts/select/read-selected/types';
import type { Choice } from '#console/types';

export const readSelectedChoiceLineMode = async <T>(
	message: string,
	choices: Array<Choice<T>>,
	defaultValue: T | undefined,
	hasDefault: boolean,
	hint?: string,
): Promise<SelectedChoiceReadResult<T>> => {
	const rendered = renderChoices(choices);

	const answer = await ask(`${message}\n${rendered}\n`, hint);

	if (answer.trim() === '' && hasDefault) {
		const choice = selectedChoiceByDefault(choices, defaultValue, hasDefault);

		if (choice) {
			return selectedChoiceResult(choice, false);
		}
	}

	const choice = findChoice(choices, answer);

	if (!choice || choice.disabled) {
		throw invalidSelectedChoice();
	}

	return selectedChoiceResult(choice, false);
};

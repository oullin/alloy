import { ask } from '#console/prompt';
import { renderChoices } from '#console/theme';
import { choicesFromCommaSeparated, defaultChoiceValues } from '#console/prompts/select/multiple';
import { multipleChoicesValueResult } from '#console/prompts/select/read-multiple/result';
import type { MultipleChoicesReadResult } from '#console/prompts/select/read-multiple/types';
import type { Choice } from '#console/types';

export const readLineMultipleChoices = async <T>(message: string, choices: Array<Choice<T>>, defaults: T[] = [], hint?: string): Promise<MultipleChoicesReadResult<T>> => {
	const rendered = renderChoices(choices);

	const answer = await ask(`${message}\n${rendered}\n`, hint);

	const value = answer.trim() === '' ? defaultChoiceValues(choices, defaults) : choicesFromCommaSeparated(choices, answer);

	return multipleChoicesValueResult(value);
};

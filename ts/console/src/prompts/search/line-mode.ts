import { choiceByValue, choiceValueEquals, findChoice } from '#console/concerns/choices';
import { parseChoiceAnswerList } from '#console/concerns/validators/choice-answer';
import { PromptValidationError } from '#console/prompt';
import { resolveSearchChoices } from '#console/prompts/search/choices';
import type { Choice, MultiSearchPromptOptions, SearchPromptOptions } from '#console/types';

type SearchLineOptions<T> = SearchPromptOptions<T> & {
	hasDefault?: boolean;
};

export const resolveLineSearchChoice = async <T>(options: SearchLineOptions<T>, query: string): Promise<T | undefined> => {
	const choices = await resolveSearchChoices(options.options, query);

	if (query === '' && options.hasDefault === true) {
		const disabledChoice = choices.find((candidate) => candidate.disabled && choiceValueEquals(candidate.value, options.default));
		const choice = choiceByValue(choices, options.default);

		return disabledChoice ? undefined : (choice?.value ?? options.default);
	}

	const matched = findChoice(choices, query);

	if (matched?.disabled) {
		return undefined;
	}

	const choice = matched ?? choices.find((candidate) => !candidate.disabled);

	return choice?.value;
};

export const resolveLineMultiSearchChoices = async <T>(options: MultiSearchPromptOptions<T>, query: string): Promise<T[]> => {
	const choices = await resolveSearchChoices(options.options, query);

	if (query === '' && options.default !== undefined) {
		return options.default.flatMap((value) => {
			const choice = choices.find((candidate) => choiceValueEquals(candidate.value, value));

			if (choice?.disabled) {
				return [];
			}

			return [choice?.value ?? value];
		});
	}

	const parts = parseChoiceAnswerList(query);

	const selectedChoices = parts.map((part) => findChoice(choices, part)).filter((choice): choice is Choice<T> => choice !== undefined && !choice.disabled);

	if (selectedChoices.length !== parts.length) {
		throw new PromptValidationError('Please select valid options.');
	}

	return selectedChoices.map((choice) => choice.value);
};

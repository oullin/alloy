import { choiceValueEquals, findChoice } from '#console/concerns/choices';
import { parseChoiceAnswerList } from '#console/concerns/validators/choice-answer';
import { PromptValidationError } from '#console/prompt';
import type { Choice } from '#console/types';

export const choicesFromCommaSeparated = <T>(choices: Array<Choice<T>>, answer: string): T[] => {
	const parts = parseChoiceAnswerList(answer);

	const selected = parts.map((part) => findChoice(choices, part)).filter((choice): choice is Choice<T> => choice !== undefined && !choice.disabled);

	if (selected.length !== parts.length) {
		throw new PromptValidationError('Please select valid options.');
	}

	return selected.map((choice) => choice.value);
};

export const markedChoiceIndexes = <T>(choices: Array<Choice<T>>, defaults: T[] = []): Set<number> => {
	return new Set(choices.flatMap((choice, index) => (!choice.disabled && defaults.some((value) => choiceValueEquals(choice.value, value)) ? [index] : [])));
};

export const markedChoiceValues = <T>(choices: Array<Choice<T>>, marked: Set<number>): T[] => {
	return [...marked].map((index) => choices[index]?.value).filter((value): value is T => value !== undefined);
};

export const defaultChoiceValues = <T>(choices: Array<Choice<T>>, defaults: T[] = []): T[] => {
	return defaults.flatMap((value) => {
		const choice = choices.find((candidate) => choiceValueEquals(candidate.value, value));

		if (choice?.disabled) {
			return [];
		}

		return [choice?.value ?? value];
	});
};

export const toggleMarkedChoice = <T>(choices: Array<Choice<T>>, marked: Set<number>, index: number): Set<number> => {
	const next = new Set(marked);

	if (next.has(index)) {
		next.delete(index);
	} else if (!choices[index]?.disabled) {
		next.add(index);
	}

	return next;
};

export const toggleAllEnabledChoices = <T>(choices: Array<Choice<T>>, marked: Set<number>): Set<number> => {
	const enabledIndexes = choices.flatMap((choice, index) => (choice.disabled ? [] : [index]));
	const allEnabledMarked = enabledIndexes.every((index) => marked.has(index));

	if (allEnabledMarked) {
		return new Set();
	}

	return new Set(enabledIndexes);
};

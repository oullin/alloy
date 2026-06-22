import { choiceValueEquals } from '#console/concerns/choices';
import type { Choice } from '#console/types';

export type SearchSelection<T> = Map<T, string>;

export const createInitialSearchSelection = <T>(choices: Array<Choice<T>>, defaults: T[] = []): SearchSelection<T> => {
	const selected = new Map<T, string>();

	for (const value of defaults) {
		const choice = choices.find((candidate) => choiceValueEquals(candidate.value, value));

		if (choice?.disabled) {
			continue;
		}

		selected.set(choice?.value ?? value, choice?.label ?? String(value));
	}

	return selected;
};

export const displayedSearchChoices = <T>(choices: Array<Choice<T>>, selected: SearchSelection<T>, query: string): Array<Choice<T>> => {
	if (query.trim() !== '') {
		return choices;
	}

	const selectedChoices = [...selected.entries()].filter(([value]) => !choices.some((choice) => choiceValueEquals(choice.value, value))).map(([value, label]) => ({ label, value }));

	return [...selectedChoices, ...choices];
};

export const selectableSearchChoices = <T>(choices: Array<Choice<T>>): Array<Choice<T>> => choices.filter((choice) => !choice.disabled);

export const markedSearchChoiceIndexes = <T>(choices: Array<Choice<T>>, selected: SearchSelection<T>): Set<number> => {
	return new Set(choices.flatMap((choice, index) => (!choice.disabled && [...selected.keys()].some((value) => choiceValueEquals(choice.value, value)) ? [index] : [])));
};

export const toggleSearchChoice = <T>(selected: SearchSelection<T>, choice: Choice<T>): void => {
	if (choice.disabled) {
		return;
	}

	const selectedValue = [...selected.keys()].find((value) => choiceValueEquals(choice.value, value));

	if (selectedValue !== undefined) {
		selected.delete(selectedValue);

		return;
	}

	selected.set(choice.value, choice.label);
};

export const toggleSearchChoices = <T>(selected: SearchSelection<T>, choices: Array<Choice<T>>): void => {
	const selectable = selectableSearchChoices(choices);
	const allSelected = selectable.every((choice) => [...selected.keys()].some((value) => choiceValueEquals(choice.value, value)));

	if (allSelected) {
		for (const choice of selectable) {
			const selectedValue = [...selected.keys()].find((value) => choiceValueEquals(choice.value, value));

			if (selectedValue !== undefined) {
				selected.delete(selectedValue);
			}
		}

		return;
	}

	for (const choice of selectable) {
		selected.set(choice.value, choice.label);
	}
};

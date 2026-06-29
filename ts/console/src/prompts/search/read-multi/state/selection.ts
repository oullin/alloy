import { createInitialSearchSelection, markedSearchChoiceIndexes, toggleSearchChoices } from '#console/prompts/search/selection';
import { toggleHighlightedSearchChoice } from '#console/prompts/search/read-multi/result';
import type { SearchSelection } from '#console/prompts/search/selection';
import type { Choice } from '#console/types';

export const createMultiSearchSelection = <T>(choices: Array<Choice<T>>, defaults: T[] | undefined): SearchSelection<T> => {
	return createInitialSearchSelection(choices, defaults);
};

export const multiSearchSelectedLabels = <T>(selected: SearchSelection<T>): string[] => {
	return [...selected.values()];
};

export const markedMultiSearchChoiceIndexes = <T>(choices: Array<Choice<T>>, selected: SearchSelection<T>): Set<number> => {
	return markedSearchChoiceIndexes(choices, selected);
};

export const toggleAllDisplayedMultiSearchChoices = <T>(selected: SearchSelection<T>, choices: Array<Choice<T>>): void => {
	toggleSearchChoices(selected, choices);
};

export const toggleHighlightedMultiSearchChoice = <T>(selected: SearchSelection<T>, choices: Array<Choice<T>>, highlighted: number | null): void => {
	toggleHighlightedSearchChoice(selected, choices, highlighted);
};

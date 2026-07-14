import type { SearchNavigationAction } from '#console/prompts/search/keys';
import type { MultiSearchChoiceQuery } from '#console/prompts/search/read-multi/choice-query';
import type { MultiSearchHighlightState } from '#console/prompts/search/read-multi/state/highlight';
import type { SearchSelection } from '#console/prompts/search/selection';
import type { Choice } from '#console/types';

import { markedMultiSearchChoiceIndexes, toggleAllDisplayedMultiSearchChoices, toggleHighlightedMultiSearchChoice } from '#console/prompts/search/read-multi/state/selection';

export const displayedMultiSearchChoices = <T>(query: MultiSearchChoiceQuery<T>): Array<Choice<T>> => {
	return query.displayedChoices();
};

export const markedDisplayedMultiSearchChoiceIndexes = <T>(query: MultiSearchChoiceQuery<T>, selected: SearchSelection<T>): Set<number> => {
	return markedMultiSearchChoiceIndexes(displayedMultiSearchChoices(query), selected);
};

export const moveDisplayedMultiSearchHighlight = async <T>(query: MultiSearchChoiceQuery<T>, highlighted: MultiSearchHighlightState<T>, action: SearchNavigationAction): Promise<void> => {
	await query.resolveChoices();

	highlighted.move(displayedMultiSearchChoices(query), action);
};

export const toggleAllDisplayedMultiSearchSelection = <T>(query: MultiSearchChoiceQuery<T>, selected: SearchSelection<T>): void => {
	toggleAllDisplayedMultiSearchChoices(selected, displayedMultiSearchChoices(query));
};

export const toggleHighlightedDisplayedMultiSearchSelection = <T>(query: MultiSearchChoiceQuery<T>, selected: SearchSelection<T>, highlighted: MultiSearchHighlightState<T>): void => {
	toggleHighlightedMultiSearchChoice(selected, displayedMultiSearchChoices(query), highlighted.value());
};

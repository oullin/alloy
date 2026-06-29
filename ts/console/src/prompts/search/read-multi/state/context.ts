import { resolveSearchChoices } from '#console/prompts/search/choices';
import { createMultiSearchChoiceQuery } from '#console/prompts/search/read-multi/choice-query';
import { createMultiSearchHighlightState } from '#console/prompts/search/read-multi/state/highlight';
import { createMultiSearchSelection } from '#console/prompts/search/read-multi/state/selection';
import type { MultiSearchChoiceQuery } from '#console/prompts/search/read-multi/choice-query';
import type { MultiSearchHighlightState } from '#console/prompts/search/read-multi/state/highlight';
import type { SearchSelection } from '#console/prompts/search/selection';
import type { MultiSearchPromptOptions } from '#console/types';

export type MultiSearchReaderStateContext<T> = {
	highlighted: MultiSearchHighlightState<T>;
	query: MultiSearchChoiceQuery<T>;
	selected: SearchSelection<T>;
};

export const createMultiSearchReaderStateContext = async <T>(options: MultiSearchPromptOptions<T>): Promise<MultiSearchReaderStateContext<T>> => {
	const initialChoices = await resolveSearchChoices(options.options, '');

	const selected = createMultiSearchSelection(initialChoices, options.default);

	return {
		highlighted: createMultiSearchHighlightState(options.scroll),
		query: createMultiSearchChoiceQuery(options, selected, initialChoices),
		selected,
	};
};

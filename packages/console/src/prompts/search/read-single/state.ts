import { resolveSearchChoices } from '#console/prompts/search/choices';
import { createSingleSearchChoiceQuery } from '#console/prompts/search/read-single/choice-query';
import { createSingleSearchHighlightState } from '#console/prompts/search/read-single/state/highlight';
import { defaultSingleSearchSelection, selectedSingleSearchSelection } from '#console/prompts/search/read-single/state/selection';
import type { SearchNavigationAction } from '#console/prompts/search/keys';
import type { SearchReaderSelection, SearchReadOptions } from '#console/prompts/search/read-single/types';
import type { TypedValueState } from '#console/typed-value/types';
import type { Choice } from '#console/types';

export type SingleSearchReaderState<T> = {
	applyTypedInput(key: string): Promise<{ cancelled: boolean }>;
	choices(): Array<Choice<T>>;
	clearHighlight(): void;
	defaultSelection(): Promise<SearchReaderSelection<T>>;
	highlighted(): number | null;
	move(action: SearchNavigationAction): Promise<void>;
	query(): TypedValueState;
	selectedSelection(): SearchReaderSelection<T>;
};

export const createSingleSearchReaderState = async <T>(options: SearchReadOptions<T>, attempt: number): Promise<SingleSearchReaderState<T>> => {
	const initialChoices = await resolveSearchChoices(options.options, '');

	const query = createSingleSearchChoiceQuery(options, initialChoices);
	const highlighted = createSingleSearchHighlightState(initialChoices, attempt, options.scroll);

	return {
		async applyTypedInput(key) {
			const next = await query.applyTypedInput(key);

			highlighted.clear();

			return next;
		},
		choices() {
			return query.choices();
		},
		clearHighlight() {
			highlighted.clear();
		},
		async defaultSelection() {
			return defaultSingleSearchSelection(options, query);
		},
		highlighted() {
			return highlighted.value();
		},
		async move(action) {
			await query.resolveChoices();

			highlighted.move(query.choices(), action);
		},
		query() {
			return query.value();
		},
		selectedSelection() {
			return selectedSingleSearchSelection(query.choices(), highlighted.value());
		},
	};
};

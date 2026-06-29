import { acceptAutocompleteMatch } from '#console/prompts/suggest/autocomplete';
import { createAutocompleteHighlightState } from '#console/prompts/suggest/read-autocomplete/state/highlight';
import { createSuggestQueryState } from '#console/prompts/suggest/state/query';
import type { SuggestOptions } from '#console/prompts/suggest/options';
import type { TypedValueState } from '#console/typed-value/types';

export type AutocompleteReaderState = {
	acceptHighlighted(requireGrowth: boolean): Promise<void>;
	applyTypedInput(key: string): Promise<{ cancelled: boolean; submitted: boolean }>;
	highlighted(): number;
	matches(): string[];
	move(direction: 1 | -1): Promise<void>;
	value(): TypedValueState;
};

export const createAutocompleteReaderState = async (options: SuggestOptions): Promise<AutocompleteReaderState> => {
	const query = await createSuggestQueryState(options);

	const highlighted = createAutocompleteHighlightState();

	return {
		async acceptHighlighted(requireGrowth) {
			await query.resolve();

			const next = acceptAutocompleteMatch(query.value(), query.matches()[highlighted.value()], requireGrowth);

			if (next !== null) {
				await query.set(next);
			} else {
				highlighted.reset();
			}
		},
		async applyTypedInput(key) {
			const next = await query.applyTypedInput(key);

			if (next.submitted || next.cancelled) {
				return { cancelled: next.cancelled, submitted: next.submitted };
			}

			if (next.changed) {
				highlighted.reset();
			}

			return { cancelled: false, submitted: false };
		},
		highlighted() {
			return highlighted.value();
		},
		matches() {
			return query.matches();
		},
		async move(direction) {
			await query.resolve();

			highlighted.move(query.matches(), direction);
		},
		value() {
			return query.value();
		},
	};
};

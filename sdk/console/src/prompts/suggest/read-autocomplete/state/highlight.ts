import { moveAutocompleteHighlight } from '#console/prompts/suggest/autocomplete';

export type AutocompleteHighlightState = {
	move(matches: string[], direction: 1 | -1): void;
	reset(): void;
	value(): number;
};

export const createAutocompleteHighlightState = (): AutocompleteHighlightState => {
	let highlighted = 0;

	return {
		move(matches, direction) {
			highlighted = moveAutocompleteHighlight(matches, highlighted, direction);
		},
		reset() {
			highlighted = 0;
		},
		value() {
			return highlighted;
		},
	};
};

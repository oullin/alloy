import { moveSuggestionHighlight } from '#console/prompts/suggest/keys';
import type { SuggestNavigationAction } from '#console/prompts/suggest/keys';

export type SuggestHighlightState = {
	clear(): void;
	move(matches: string[], action: SuggestNavigationAction, scroll?: number): void;
	value(): number | null;
};

export const createSuggestHighlightState = (): SuggestHighlightState => {
	let highlighted: number | null = null;

	return {
		clear() {
			highlighted = null;
		},
		move(matches, action, scroll) {
			highlighted = moveSuggestionHighlight(matches, highlighted, action, scroll);
		},
		value() {
			return highlighted;
		},
	};
};

import { pageIndex } from '#console/prompts/select/navigation';

export const nextSuggestionHighlight = (matches: string[], highlighted: number | null, direction: 1 | -1): number | null => {
	if (matches.length === 0) {
		return null;
	}

	const current = highlighted ?? (direction === 1 ? -1 : matches.length);

	return (current + direction + matches.length) % matches.length;
};

export const pageSuggestionHighlight = (matches: string[], highlighted: number | null, direction: 1 | -1, scroll?: number): number | null => {
	if (matches.length === 0) {
		return null;
	}

	if (highlighted === null) {
		return direction === 1 ? 0 : Math.max(0, matches.length - 1);
	}

	return pageIndex(matches.length, highlighted, direction, scroll);
};

export const firstSuggestionHighlight = (matches: string[]): number | null => {
	return matches.length === 0 ? null : 0;
};

export const lastSuggestionHighlight = (matches: string[]): number | null => {
	return matches.length === 0 ? null : matches.length - 1;
};

export const nextAutocompleteHighlight = (matches: string[], highlighted: number, direction: 1 | -1): number => {
	return matches.length === 0 ? 0 : (highlighted + direction + matches.length) % matches.length;
};

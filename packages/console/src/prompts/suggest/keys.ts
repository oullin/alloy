import { Key, oneOf } from '#console/key';
import { firstSuggestionHighlight, lastSuggestionHighlight, nextSuggestionHighlight, pageSuggestionHighlight } from '#console/prompts/suggest/navigation';

export type SuggestNavigationAction = 'first' | 'last' | 'next' | 'page-next' | 'page-previous' | 'previous';

export const suggestNavigationAction = (key: string): SuggestNavigationAction | null => {
	if (key === Key.tab || key === Key.down || key === Key.downArrow || key === Key.ctrlN) {
		return 'next';
	}

	if (key === Key.up || key === Key.upArrow || key === Key.ctrlP || key === Key.shiftTab) {
		return 'previous';
	}

	if (key === Key.pageDown) {
		return 'page-next';
	}

	if (key === Key.pageUp) {
		return 'page-previous';
	}

	if (oneOf([Key.home, Key.ctrlA], key)) {
		return 'first';
	}

	if (oneOf([Key.end, Key.ctrlE], key)) {
		return 'last';
	}

	return null;
};

export const clearsSuggestionHighlight = (key: string): boolean => oneOf([Key.left, Key.leftArrow, Key.right, Key.rightArrow, Key.ctrlB, Key.ctrlF], key) !== undefined;

export const moveSuggestionHighlight = (matches: string[], highlighted: number | null, action: SuggestNavigationAction, scroll?: number): number | null => {
	if (action === 'next') {
		return nextSuggestionHighlight(matches, highlighted, 1);
	}

	if (action === 'previous') {
		return nextSuggestionHighlight(matches, highlighted, -1);
	}

	if (action === 'page-next') {
		return pageSuggestionHighlight(matches, highlighted, 1, scroll);
	}

	if (action === 'page-previous') {
		return pageSuggestionHighlight(matches, highlighted, -1, scroll);
	}

	if (action === 'first') {
		return firstSuggestionHighlight(matches);
	}

	return lastSuggestionHighlight(matches);
};

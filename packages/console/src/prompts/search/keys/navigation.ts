import { Key, oneOf } from '#console/key';

export type SearchNavigationAction = 'first' | 'last' | 'next' | 'page-next' | 'page-previous' | 'previous';

export type SearchNavigationOptions = {
	controlNavigation?: boolean;
	lineControls?: boolean;
};

export const searchNavigationAction = (key: string, options: SearchNavigationOptions = {}): SearchNavigationAction | null => {
	if (key === Key.down || key === Key.downArrow || key === Key.tab || (options.controlNavigation === true && key === Key.ctrlN)) {
		return 'next';
	}

	if (key === Key.up || key === Key.upArrow || key === Key.shiftTab || (options.controlNavigation === true && key === Key.ctrlP)) {
		return 'previous';
	}

	if (key === Key.pageDown) {
		return 'page-next';
	}

	if (key === Key.pageUp) {
		return 'page-previous';
	}

	if (oneOf(options.lineControls === true ? [Key.home, Key.ctrlA] : [Key.home], key)) {
		return 'first';
	}

	if (oneOf(options.lineControls === true ? [Key.end, Key.ctrlE] : [Key.end], key)) {
		return 'last';
	}

	return null;
};

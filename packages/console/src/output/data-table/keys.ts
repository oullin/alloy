import { Key, oneOf } from '#console/key';

export type DataTableNavigationAction = 'first' | 'last' | 'next' | 'page-next' | 'page-previous' | 'previous';

export const dataTableNavigationAction = (key: string): DataTableNavigationAction | null => {
	if (key === Key.down || key === Key.downArrow || key === Key.ctrlN || key === Key.tab) {
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

export const startsDataTableSearch = (key: string): boolean => key === '/';

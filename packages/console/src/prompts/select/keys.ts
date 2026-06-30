import { Key, oneOf } from '#console/key';
import { firstEnabledIndex, nextEnabledIndex } from '#console/concerns/choices';
import { lastEnabledChoiceIndex, nextChoiceKeys, pageEnabledChoiceIndex, previousChoiceKeys } from '#console/prompts/select/navigation';
import type { Choice } from '#console/types';

export type SelectNavigationAction = 'first' | 'last' | 'next' | 'page-next' | 'page-previous' | 'previous';

type SelectNavigationOptions = {
	lineControls?: boolean;
};

export const selectNavigationAction = (key: string, options: SelectNavigationOptions = {}): SelectNavigationAction | null => {
	if (nextChoiceKeys(key)) {
		return 'next';
	}

	if (previousChoiceKeys(key)) {
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

export const moveSelectHighlight = <T>(choices: Array<Choice<T>>, selected: number, action: SelectNavigationAction, scroll?: number): number => {
	if (action === 'next') {
		return nextEnabledIndex(choices, selected, 1);
	}

	if (action === 'previous') {
		return nextEnabledIndex(choices, selected, -1);
	}

	if (action === 'page-next') {
		return pageEnabledChoiceIndex(choices, selected, 1, scroll);
	}

	if (action === 'page-previous') {
		return pageEnabledChoiceIndex(choices, selected, -1, scroll);
	}

	if (action === 'first') {
		return firstEnabledIndex(choices);
	}

	return lastEnabledChoiceIndex(choices);
};

import { Key, oneOf } from '#console/key';
import { firstSearchHighlight, lastSearchHighlight, nextRetriedSearchHighlight, nextSearchHighlight, pageSearchHighlight } from '#console/prompts/search/navigation';
import type { SearchNavigationAction } from '#console/prompts/search/keys/navigation';
import type { Choice } from '#console/types';

export type MoveSearchHighlightOptions = {
	attempt?: number;
	retryFirst?: boolean;
	scroll?: number;
};

export const clearsSearchHighlight = (key: string): boolean => oneOf([Key.left, Key.leftArrow, Key.right, Key.rightArrow, Key.ctrlB, Key.ctrlF], key) !== undefined;

export const moveSearchHighlight = <T>(choices: Array<Choice<T>>, highlighted: number | null, action: SearchNavigationAction, options: MoveSearchHighlightOptions = {}): number | null => {
	if (action === 'next') {
		return options.retryFirst === true ? nextRetriedSearchHighlight(choices, highlighted, options.attempt ?? 0) : nextSearchHighlight(choices, highlighted, 1);
	}

	if (action === 'previous') {
		return nextSearchHighlight(choices, highlighted, -1);
	}

	if (action === 'page-next') {
		return pageSearchHighlight(choices, highlighted, 1, options.scroll);
	}

	if (action === 'page-previous') {
		return pageSearchHighlight(choices, highlighted, -1, options.scroll);
	}

	if (action === 'first') {
		return firstSearchHighlight(choices);
	}

	return lastSearchHighlight(choices);
};

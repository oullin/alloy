import { firstEnabledIndex, nextEnabledIndex } from '#console/concerns/choices';
import { lastEnabledIndex } from '#console/prompts/search/choices';
import { pageEnabledChoiceIndex } from '#console/prompts/select/navigation';
import type { Choice } from '#console/types';

export const firstSearchHighlight = <T>(choices: Array<Choice<T>>): number | null => {
	return choices.length === 0 ? null : firstEnabledIndex(choices);
};

export const initialRetriedSearchHighlight = <T>(choices: Array<Choice<T>>, attempt: number): number | null => {
	return attempt > 0 && choices.length > 0 ? 0 : null;
};

export const lastSearchHighlight = <T>(choices: Array<Choice<T>>): number | null => {
	return choices.length === 0 ? null : lastEnabledIndex(choices);
};

export const nextSearchHighlight = <T>(choices: Array<Choice<T>>, highlighted: number | null, direction: 1 | -1): number | null => {
	if (choices.length === 0) {
		return null;
	}

	return highlighted === null ? (direction === 1 ? firstEnabledIndex(choices) : lastEnabledIndex(choices)) : nextEnabledIndex(choices, highlighted, direction);
};

export const nextRetriedSearchHighlight = <T>(choices: Array<Choice<T>>, highlighted: number | null, attempt: number): number | null => {
	if (choices.length === 0) {
		return null;
	}

	if (highlighted !== null) {
		return nextEnabledIndex(choices, highlighted, 1);
	}

	const first = firstEnabledIndex(choices);

	return attempt > 0 ? nextEnabledIndex(choices, first, 1) : first;
};

export const pageSearchHighlight = <T>(choices: Array<Choice<T>>, highlighted: number | null, direction: 1 | -1, scroll?: number): number | null => {
	if (choices.length === 0) {
		return null;
	}

	return highlighted === null ? (direction === 1 ? firstEnabledIndex(choices) : lastEnabledIndex(choices)) : pageEnabledChoiceIndex(choices, highlighted, direction, scroll);
};

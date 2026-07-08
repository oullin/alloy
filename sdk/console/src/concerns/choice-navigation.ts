import { parseOptionalScrollSize } from '#console/concerns/validators/scroll';
import type { Choice } from '#console/types';

export const firstEnabledIndex = <T>(choices: Array<Choice<T>>): number => {
	const index = choices.findIndex((choice) => !choice.disabled);

	return index === -1 ? 0 : index;
};

export const nextEnabledIndex = <T>(choices: Array<Choice<T>>, current: number, direction: 1 | -1): number => {
	if (choices.length === 0) {
		return 0;
	}

	let index = current;

	for (let attempts = 0; attempts < choices.length; attempts += 1) {
		index = (index + direction + choices.length) % choices.length;

		if (!choices[index]?.disabled) {
			return index;
		}
	}

	return current;
};

export const choiceWindow = (total: number, selected: number, scroll?: number): { end: number; start: number } => {
	const size = parseOptionalScrollSize(scroll);

	if (size === undefined || size >= total) {
		return { end: total, start: 0 };
	}

	const before = Math.floor((size - 1) / 2);
	const start = Math.max(0, Math.min(selected - before, total - size));

	return { end: start + size, start };
};

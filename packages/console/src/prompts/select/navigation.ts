import { parseChoiceAnswerIndex } from '#console/concerns/validators/choice-answer';
import { Key } from '#console/key';
import { parseScrollSize } from '#console/concerns/validators/scroll';
import type { Choice } from '#console/types';

export const parseChoiceIndex = (key: string): number | null => parseChoiceAnswerIndex(key);

export const previousChoiceKeys = (key: string): boolean => {
	return key === Key.up || key === Key.upArrow || key === Key.left || key === Key.leftArrow || key === Key.shiftTab || key === Key.ctrlP || key === Key.ctrlB || key === 'k' || key === 'h';
};

export const nextChoiceKeys = (key: string): boolean => {
	return key === Key.down || key === Key.downArrow || key === Key.right || key === Key.rightArrow || key === Key.tab || key === Key.ctrlN || key === Key.ctrlF || key === 'j' || key === 'l';
};

export const lastEnabledChoiceIndex = <T>(choices: Array<Choice<T>>): number => {
	let selected = choices.length - 1;

	while (choices[selected]?.disabled && selected > 0) {
		selected -= 1;
	}

	return selected;
};

export const pageSize = (scroll?: number): number => {
	return parseScrollSize(scroll, 10);
};

export const pageIndex = (total: number, current: number, direction: 1 | -1, scroll?: number): number => {
	if (total <= 0) {
		return 0;
	}

	return Math.max(0, Math.min(total - 1, current + pageSize(scroll) * direction));
};

export const pageEnabledChoiceIndex = <T>(choices: Array<Choice<T>>, current: number, direction: 1 | -1, scroll?: number): number => {
	const target = pageIndex(choices.length, current, direction, scroll);

	for (let index = target; index >= 0 && index < choices.length; index += direction) {
		if (!choices[index]?.disabled) {
			return index;
		}
	}

	for (let index = target; index >= 0 && index < choices.length; index -= direction) {
		if (!choices[index]?.disabled) {
			return index;
		}
	}

	return current;
};

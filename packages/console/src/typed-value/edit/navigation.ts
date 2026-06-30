import { Key, oneOf } from '#console/key';
import { moveLine, moveToLineBoundary } from '#console/typed-value/lines';

export const moveTypedValueCursor = (value: string[], cursor: number, key: string, allowNewLine: boolean, wrapWidth?: number): number | undefined => {
	if (key === Key.left || key === Key.leftArrow || key === Key.ctrlB) {
		return Math.max(0, cursor - 1);
	}

	if (key === Key.right || key === Key.rightArrow || key === Key.ctrlF) {
		return Math.min(value.length, cursor + 1);
	}

	if (allowNewLine && (key === Key.up || key === Key.upArrow || key === Key.ctrlP)) {
		return moveLine(value, cursor, -1, wrapWidth);
	}

	if (allowNewLine && (key === Key.down || key === Key.downArrow || key === Key.ctrlN)) {
		return moveLine(value, cursor, 1, wrapWidth);
	}

	if (oneOf([Key.home, Key.ctrlA], key)) {
		return allowNewLine ? moveToLineBoundary(value, cursor, 'start', wrapWidth) : 0;
	}

	if (oneOf([Key.end, Key.ctrlE], key)) {
		return allowNewLine ? moveToLineBoundary(value, cursor, 'end', wrapWidth) : value.length;
	}

	return undefined;
};

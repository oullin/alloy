import { characterLength, characters } from '#console/typed-value/characters';
import { lineRanges } from '#console/typed-value/line-ranges';
import { visibleLineWindow } from '#console/typed-value/lines/window';
import { parseTypedValueCursor } from '#console/typed-value/validators/cursor';
import type { VisibleTextWindow } from '#console/typed-value/lines/types';

export const visibleTextWindow = (value: string, cursor: number, rows: number | undefined, width?: number): VisibleTextWindow => {
	const valueCharacters = characters(value);
	const visibleCursor = parseTypedValueCursor(cursor, valueCharacters.length);
	const ranges = lineRanges(valueCharacters, width);
	const window = visibleLineWindow(value, visibleCursor, rows, width);
	const visibleRanges = ranges.slice(window.start, window.start + window.lines.length);

	const currentRangeIndex = Math.max(
		0,
		visibleRanges.findIndex((range) => visibleCursor <= range.end),
	);

	const currentRange = visibleRanges[currentRangeIndex] ?? visibleRanges.at(-1);
	const cursorInLine = currentRange === undefined ? 0 : Math.max(0, Math.min(visibleCursor, currentRange.end) - currentRange.start);
	const previousWidth = window.lines.slice(0, currentRangeIndex).reduce((width, line) => width + characterLength(line) + 1, 0);

	return {
		...window,
		cursor: previousWidth + cursorInLine,
		text: window.lines.join('\n'),
	};
};

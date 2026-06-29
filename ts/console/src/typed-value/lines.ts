import { currentLine, lineRanges } from '#console/typed-value/line-ranges';
import { visibleLineWindow } from '#console/typed-value/lines/window';

export { visibleLineWindow } from '#console/typed-value/lines/window';
export { visibleTextWindow } from '#console/typed-value/lines/text-window';
export type { VisibleLineWindow, VisibleTextWindow } from '#console/typed-value/lines/types';

export const visibleLines = (value: string, cursor: number, rows: number | undefined): string => {
	return visibleLineWindow(value, cursor, rows).lines.join('\n');
};

export const moveLine = (value: string[], cursor: number, direction: 1 | -1, width?: number): number => {
	const ranges = lineRanges(value, width);
	const index = currentLine(ranges, cursor);
	const range = ranges[index];

	if (!range) {
		return cursor;
	}

	const target = ranges[index + direction];

	if (!target) {
		return direction === -1 ? 0 : value.length;
	}

	const column = Math.min(cursor - range.start, range.end - range.start);
	const targetColumn = Math.min(column, target.end - target.start);

	return target.start + targetColumn;
};

export const moveToLineBoundary = (value: string[], cursor: number, boundary: 'start' | 'end', width?: number): number => {
	const ranges = lineRanges(value, width);
	const range = ranges[currentLine(ranges, cursor)];

	if (!range) {
		return cursor;
	}

	return boundary === 'start' ? range.start : range.end;
};

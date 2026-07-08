import { characters, fromCharacters } from '#console/typed-value/characters';
import { currentLine, lineRanges } from '#console/typed-value/line-ranges';
import { parseTypedValueCursor } from '#console/typed-value/validators/cursor';
import { parseTypedValueRows } from '#console/typed-value/validators/rows';
import type { VisibleLineWindow } from '#console/typed-value/lines/types';

export const visibleLineWindow = (value: string, cursor: number, rows: number | undefined, width?: number): VisibleLineWindow => {
	const valueCharacters = characters(value);
	const ranges = lineRanges(valueCharacters, width);
	const visibleCursor = parseTypedValueCursor(cursor, valueCharacters.length);
	const visibleRows = parseTypedValueRows(rows);

	if (visibleRows === undefined) {
		return {
			lines: ranges.map((range) => fromCharacters(valueCharacters.slice(range.start, range.end))),
			start: 0,
			total: ranges.length,
		};
	}

	const line = currentLine(ranges, visibleCursor);
	const start = Math.max(0, line - visibleRows + 1);
	const end = start + visibleRows;

	return {
		lines: ranges.slice(start, end).map((range) => fromCharacters(valueCharacters.slice(range.start, range.end))),
		start,
		total: ranges.length,
	};
};

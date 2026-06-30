import { visibleWidth } from '#console/strings';
import { parseTypedValueWrapWidth } from '#console/typed-value/validators/width';

export type LineRange = {
	end: number;
	start: number;
};

export const lineRanges = (value: string[], width?: number): LineRange[] => {
	const ranges: LineRange[] = [];
	const wrapWidth = parseTypedValueWrapWidth(width);

	let start = 0;

	for (const [index, character] of value.entries()) {
		if (character === '\n') {
			appendRanges(ranges, wrappedRanges(value, start, index, wrapWidth));
			start = index + 1;
		}
	}

	appendRanges(ranges, wrappedRanges(value, start, value.length, wrapWidth));

	return ranges;
};

export const currentLine = (ranges: LineRange[], cursor: number): number => {
	const index = ranges.findIndex((range) => cursor <= range.end);

	return index === -1 ? ranges.length - 1 : index;
};

const appendRanges = (ranges: LineRange[], nextRanges: LineRange[]): void => {
	for (const range of nextRanges) {
		ranges.push(range);
	}
};

const wrappedRanges = (value: string[], start: number, end: number, width?: number): LineRange[] => {
	if (width === undefined || width <= 0 || start === end) {
		return [{ end, start }];
	}

	const ranges: LineRange[] = [];

	let rangeStart = start;
	let rangeWidth = 0;

	for (let index = start; index < end; index += 1) {
		const characterWidth = visibleWidth(value[index] ?? '');

		if (rangeWidth > 0 && rangeWidth + characterWidth > width) {
			ranges.push({ end: index, start: rangeStart });
			rangeStart = index;
			rangeWidth = 0;
		}

		rangeWidth += characterWidth;
	}

	ranges.push({ end, start: rangeStart });

	return ranges;
};

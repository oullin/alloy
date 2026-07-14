import { ansiCloseSequence, parseAnsiSegments } from '#console/string-utils/ansi';
import { parseStringWidth } from '#console/string-utils/validators/width';
import { visibleWidth } from '#console/string-utils/width';

/** Truncates a string to the given visible width, appending a marker and keeping ANSI styling intact. */
export const truncate = (value: string, width: number, marker = '...'): string => {
	const parsedWidth = parseStringWidth(width);

	if (parsedWidth === undefined) {
		return '';
	}

	const markerWidth = visibleWidth(marker);

	if (visibleWidth(value) <= parsedWidth) {
		return value;
	}

	if (parsedWidth <= markerWidth) {
		let clippedMarker = '';
		let clippedWidth = 0;

		for (const char of marker) {
			const charWidth = visibleWidth(char);

			if (clippedWidth + charWidth > parsedWidth) {
				break;
			}

			clippedMarker += char;
			clippedWidth += charWidth;
		}

		return clippedMarker;
	}

	let result = '';
	let activeCodes = '';
	let resultWidth = 0;

	for (const segment of parseAnsiSegments(value)) {
		if (segment.codes !== activeCodes) {
			if (activeCodes !== '') {
				result += ansiCloseSequence(activeCodes);
			}

			if (segment.codes !== '') {
				result += segment.codes;
			}

			activeCodes = segment.codes;
		}

		for (const char of segment.text) {
			const charWidth = visibleWidth(char);

			if (resultWidth + charWidth + markerWidth > parsedWidth) {
				return `${result}${activeCodes === '' ? '' : ansiCloseSequence(activeCodes)}${marker}`;
			}

			result += char;
			resultWidth += charWidth;
		}
	}

	return result;
};

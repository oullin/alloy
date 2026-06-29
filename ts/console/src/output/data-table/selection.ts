import { parseOptionalScrollSize } from '#console/concerns/validators/scroll';
import { parseDataTableRowTotal, parseDataTableSelectedIndex } from '#console/output/data-table/validators/selection';
import type { VisibleDataTableRow } from '#console/output/data-table/types';

export const clampDataTableSelection = <T>(selected: number, rows: Array<VisibleDataTableRow<T>>): number => {
	return parseDataTableSelectedIndex(selected, rows.length);
};

export const dataTableRowWindow = (total: number, selected: number, scroll?: number): { end: number; start: number } => {
	const rowTotal = parseDataTableRowTotal(total);
	const selectedIndex = parseDataTableSelectedIndex(selected, rowTotal);
	const size = parseOptionalScrollSize(scroll);

	if (size === undefined || size >= rowTotal) {
		return { end: rowTotal, start: 0 };
	}

	const before = Math.floor((size - 1) / 2);
	const start = Math.max(0, Math.min(selectedIndex - before, rowTotal - size));

	return { end: start + size, start };
};

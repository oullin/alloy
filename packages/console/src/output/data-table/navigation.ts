import { parseScrollSize } from '#console/concerns/validators/scroll';
import { parseDataTableRowTotal, parseDataTableSelectedIndex } from '#console/output/data-table/validators/selection';
import type { DataTableNavigationAction } from '#console/output/data-table/keys';

export const nextDataTableSelection = (selected: number, total: number): number => {
	const rowTotal = parseDataTableRowTotal(total);

	return rowTotal === 0 ? 0 : (parseDataTableSelectedIndex(selected, rowTotal) + 1) % rowTotal;
};

export const previousDataTableSelection = (selected: number, total: number): number => {
	const rowTotal = parseDataTableRowTotal(total);

	return rowTotal === 0 ? 0 : (parseDataTableSelectedIndex(selected, rowTotal) - 1 + rowTotal) % rowTotal;
};

export const pageDataTableSelection = (selected: number, total: number, direction: 1 | -1, scroll?: number): number => {
	const rowTotal = parseDataTableRowTotal(total);

	if (rowTotal === 0) {
		return 0;
	}

	return Math.max(0, Math.min(rowTotal - 1, parseDataTableSelectedIndex(selected, rowTotal) + parseScrollSize(scroll, 10) * direction));
};

export const firstDataTableSelection = (): number => 0;

export const lastDataTableSelection = (total: number): number => Math.max(0, parseDataTableRowTotal(total) - 1);

export const moveDataTableSelection = (action: DataTableNavigationAction, selected: number, total: number, scroll?: number): number => {
	if (action === 'next') {
		return nextDataTableSelection(selected, total);
	}

	if (action === 'previous') {
		return previousDataTableSelection(selected, total);
	}

	if (action === 'page-next') {
		return pageDataTableSelection(selected, total, 1, scroll);
	}

	if (action === 'page-previous') {
		return pageDataTableSelection(selected, total, -1, scroll);
	}

	if (action === 'first') {
		return firstDataTableSelection();
	}

	return lastDataTableSelection(total);
};

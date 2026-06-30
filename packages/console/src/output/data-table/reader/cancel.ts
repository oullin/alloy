import { cancelPrompt } from '#console/prompt';
import { eraseRenderedFrame } from '#console/status/frame';
import { renderCancelledDataTableFrame } from '#console/output/data-table/render';
import { dataTableRowValue } from '#console/output/data-table/rows';
import { selectedDataTableValue } from '#console/output/data-table/reader/result';
import type { DataTableReadOptions } from '#console/output/data-table/reader/types';
import type { DataTableSearchMode } from '#console/output/data-table/search';
import type { DataTableSelectionReadResult, VisibleDataTableRow } from '#console/output/data-table/types';

const cancelledDataTableValue = <T>(options: DataTableReadOptions<T>, rows: Array<VisibleDataTableRow<T>>, selected: number): T | number => {
	if (rows[selected] !== undefined) {
		return selectedDataTableValue(rows, selected);
	}

	if (options.hasDefault === true && options.default !== undefined) {
		return options.default;
	}

	return dataTableRowValue(options.rows[0], 0);
};

export const cancelDataTableSelection = async <T>(
	options: DataTableReadOptions<T>,
	headers: string[],
	rows: Array<VisibleDataTableRow<T>>,
	selected: number,
	frame: string,
	mode: DataTableSearchMode,
	query: string,
): Promise<DataTableSelectionReadResult<T>> => {
	const value = cancelledDataTableValue(options, rows, selected);

	eraseRenderedFrame(frame);
	renderCancelledDataTableFrame(options.message, headers, rows, selected, mode, query);

	return {
		cancelled: true,
		rows,
		selected,
		submitted: false,
		value: await cancelPrompt(value),
	};
};

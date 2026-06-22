import { ask } from '#console/prompt/ask';
import { promptEnvironment } from '#console/environment';
import { visibleDataTableRows } from '#console/output/data-table/rows';
import { assertSelectedDataTableRow, dataTableSelectionResult, initialDataTableSelection } from '#console/output/data-table/reader/result';
import type { DataTableReadOptions } from '#console/output/data-table/reader/types';
import type { DataTableSelectionReadResult } from '#console/output/data-table/types';

export const readDataTableFallbackSelection = async <T>(options: DataTableReadOptions<T>, headers: string[]): Promise<DataTableSelectionReadResult<T>> => {
	const environment = promptEnvironment();

	const answer = environment.input.readLine ? await ask(options.message, options.hint) : '';

	const query = answer.trim();

	const rows = visibleDataTableRows(options, headers, query);
	const selected = query === '' ? initialDataTableSelection(rows, options.default, options.hasDefault) : 0;

	assertSelectedDataTableRow(rows, selected, '');

	return dataTableSelectionResult(rows, selected, false);
};

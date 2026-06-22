import { promptEnvironment } from '#console/environment';
import { parseTableOptions, tableRowCells } from '#console/output/validators/table';
import { renderTable } from '#console/theme';
import type { TableOptions } from '#console/types';

export const table = (headersOrOptions: TableOptions | string[] | TableOptions['rows'] = [], rows: TableOptions['rows'] | null = null): void => {
	const options = parseTableOptions(headersOrOptions, rows);
	const normalizedRows = options.rows.map((row) => tableRowCells(row, options.headers));

	promptEnvironment().output.write(`${renderTable(options.headers, normalizedRows)}\n`);
};

export const dataTable = table;

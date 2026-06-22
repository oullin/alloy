import { dataTableRowCells } from '#console/output/data-table/rows';
import type { VisibleDataTableRow } from '#console/output/data-table/types';

export const renderSubmittedDataTableRow = <T>(headers: string[], rows: Array<VisibleDataTableRow<T>>, selected: number): string | null => {
	const row = rows[selected];

	return row ? dataTableRowCells(headers, row.row).join(', ') : null;
};

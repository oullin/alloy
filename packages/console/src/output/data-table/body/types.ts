import type { VisibleDataTableRow } from '#console/output/data-table/types';
import type { DataTableRow } from '#console/types';

export type DataTableBodyOptions<T> = {
	allRows: Array<DataTableRow<T>>;
	headers: string[];
	query: string;
	rows: Array<VisibleDataTableRow<T>>;
	scroll?: number;
	selected: number;
};

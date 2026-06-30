import type { DataTableRow } from '#console/types';

export type VisibleDataTableRow<T> = {
	index: number;
	row: DataTableRow<T>;
};

export type DataTableSelectionReadResult<T> = {
	cancelled: boolean;
	frame?: string;
	rows: Array<VisibleDataTableRow<T>>;
	selected: number;
	submitted: boolean;
	value: T | number;
};

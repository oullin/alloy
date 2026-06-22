import { applyDataTableSearchKey, startDataTableSearch } from '#console/output/data-table/search';
import type { DataTableSearchChange, DataTableSearchState } from '#console/output/data-table/search';

export type DataTableReaderSearchChange = DataTableSearchChange;

export const applyDataTableReaderSearchKey = (state: DataTableSearchState, key: string): DataTableReaderSearchChange => {
	return applyDataTableSearchKey(state, key);
};

export const beginDataTableReaderSearch = (): DataTableSearchState => {
	return startDataTableSearch();
};

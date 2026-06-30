import { createDataTableReaderSelectionState } from '#console/output/data-table/reader/state/selection';
import { applyDataTableReaderSearchKey, beginDataTableReaderSearch } from '#console/output/data-table/reader/state/search';
import { visibleDataTableRows } from '#console/output/data-table/rows';
import { initialDataTableSearchState } from '#console/output/data-table/search';
import type { DataTableNavigationAction } from '#console/output/data-table/keys';
import type { DataTableSearchState } from '#console/output/data-table/search';
import type { VisibleDataTableRow } from '#console/output/data-table/types';
import type { DataTableReadOptions } from '#console/output/data-table/reader/types';

export type DataTableReaderState<T> = {
	applySearchKey(key: string): boolean;
	beginSearch(): void;
	mode(): DataTableSearchState['mode'];
	moveSelection(action: DataTableNavigationAction): void;
	query(): string;
	rows(): Array<VisibleDataTableRow<T>>;
	selected(): number;
	setSelected(selected: number): void;
};

export const createDataTableReaderState = <T>(options: DataTableReadOptions<T>, headers: string[]): DataTableReaderState<T> => {
	let search = initialDataTableSearchState();

	const selected = createDataTableReaderSelectionState(visibleRows(), options);

	function visibleRows(): Array<VisibleDataTableRow<T>> {
		return visibleDataTableRows(options, headers, search.query.value);
	}

	function applySearchChange(nextSearch: DataTableSearchState, resetSelection: boolean): void {
		search = nextSearch;

		if (resetSelection) {
			selected.reset();
		}
	}

	return {
		applySearchKey(key) {
			const nextSearch = applyDataTableReaderSearchKey(search, key);

			if (!nextSearch.changed) {
				return false;
			}

			applySearchChange(nextSearch.state, nextSearch.resetSelection);

			return true;
		},
		beginSearch() {
			applySearchChange(beginDataTableReaderSearch(), false);
		},
		mode() {
			return search.mode;
		},
		moveSelection(action) {
			selected.move(action, visibleRows().length);
		},
		query() {
			return search.query.value;
		},
		rows: visibleRows,
		selected() {
			return selected.value();
		},
		setSelected(nextSelected) {
			selected.set(nextSelected);
		},
	};
};

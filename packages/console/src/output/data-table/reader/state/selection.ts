import { moveDataTableSelection } from '#console/output/data-table/navigation';
import { initialDataTableSelection } from '#console/output/data-table/reader/result';
import type { DataTableNavigationAction } from '#console/output/data-table/keys';
import type { DataTableReadOptions } from '#console/output/data-table/reader/types';
import type { VisibleDataTableRow } from '#console/output/data-table/types';

export type DataTableReaderSelectionState = {
	move(action: DataTableNavigationAction, total: number): void;
	reset(): void;
	set(selected: number): void;
	value(): number;
};

export const createDataTableReaderSelectionState = <T>(rows: Array<VisibleDataTableRow<T>>, options: DataTableReadOptions<T>): DataTableReaderSelectionState => {
	let selected = initialDataTableSelection(rows, options.default, options.hasDefault);

	return {
		move(action, total) {
			selected = moveDataTableSelection(action, selected, total, options.scroll);
		},
		reset() {
			selected = 0;
		},
		set(nextSelected) {
			selected = nextSelected;
		},
		value() {
			return selected;
		},
	};
};

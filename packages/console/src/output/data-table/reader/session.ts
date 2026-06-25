import { renderDataTableReaderSessionFrame } from '#console/output/data-table/reader/session/render';
import { createDataTableReaderState } from '#console/output/data-table/reader/state';
import type { DataTableNavigationAction } from '#console/output/data-table/keys';
import type { DataTableReadOptions } from '#console/output/data-table/reader/types';
import type { DataTableSearchMode } from '#console/output/data-table/search';
import type { VisibleDataTableRow } from '#console/output/data-table/types';

export type DataTableReaderSession<T> = {
	applySearchKey(key: string): boolean;
	beginSearch(): void;
	frame(): string;
	mode(): DataTableSearchMode;
	moveSelection(action: DataTableNavigationAction): void;
	query(): string;
	render(): void;
	rows(): Array<VisibleDataTableRow<T>>;
	selected(): number;
};

export const createDataTableReaderSession = <T>(options: DataTableReadOptions<T>, headers: string[]): DataTableReaderSession<T> => {
	const state = createDataTableReaderState(options, headers);

	let frame = '';

	function render(): void {
		frame = renderDataTableReaderSessionFrame(options, headers, state, frame);
	}

	return {
		applySearchKey(key) {
			if (!state.applySearchKey(key)) {
				return false;
			}

			render();

			return true;
		},
		beginSearch() {
			state.beginSearch();
			render();
		},
		frame() {
			return frame;
		},
		mode() {
			return state.mode();
		},
		moveSelection(action) {
			state.moveSelection(action);
			render();
		},
		query() {
			return state.query();
		},
		render,
		rows: () => state.rows(),
		selected() {
			return state.selected();
		},
	};
};

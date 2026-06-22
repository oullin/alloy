import { Key } from '#console/key';
import { applyTypedKey } from '#console/typed-value';
import type { TypedValueState } from '#console/typed-value/types';

export type DataTableSearchMode = 'browse' | 'search';

export type DataTableSearchState = {
	mode: DataTableSearchMode;
	query: TypedValueState;
};

export type DataTableSearchChange = {
	changed: boolean;
	resetSelection: boolean;
	state: DataTableSearchState;
};

export const initialDataTableSearchState = (): DataTableSearchState => ({
	mode: 'browse',
	query: { cursor: 0, value: '' },
});

export const startDataTableSearch = (): DataTableSearchState => ({
	mode: 'search',
	query: { cursor: 0, value: '' },
});

export const confirmDataTableSearch = (state: DataTableSearchState): DataTableSearchState => ({
	...state,
	mode: 'browse',
});

export const clearDataTableSearch = (): DataTableSearchState => ({
	mode: 'browse',
	query: { cursor: 0, value: '' },
});

export const applyDataTableSearchKey = (state: DataTableSearchState, key: string): DataTableSearchChange => {
	if (state.mode !== 'search') {
		return { changed: false, resetSelection: false, state };
	}

	if (key === Key.enter) {
		return { changed: true, resetSelection: false, state: confirmDataTableSearch(state) };
	}

	if (key === Key.escape) {
		return { changed: true, resetSelection: true, state: clearDataTableSearch() };
	}

	const next = applyTypedKey(state.query, key);

	if (next.value === state.query.value && next.cursor === state.query.cursor) {
		return { changed: false, resetSelection: false, state };
	}

	return {
		changed: true,
		resetSelection: true,
		state: {
			mode: 'search',
			query: { cursor: next.cursor, value: next.value },
		},
	};
};

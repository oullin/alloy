import { promptEnvironment } from '#console/environment';
import { readDataTableFallbackSelection } from '#console/output/data-table/reader/fallback';
import { readDataTableSelectionInteractive } from '#console/output/data-table/reader/interactive';
import type { DataTableReadOptions } from '#console/output/data-table/reader/types';
import type { DataTableSelectionReadResult } from '#console/output/data-table/types';

export const readDataTableSelection = async <T>(options: DataTableReadOptions<T>, headers: string[]): Promise<DataTableSelectionReadResult<T>> => {
	const environment = promptEnvironment();

	if (!environment.input.readKey) {
		return readDataTableFallbackSelection(options, headers);
	}

	return readDataTableSelectionInteractive(environment.input.readKey, options, headers);
};

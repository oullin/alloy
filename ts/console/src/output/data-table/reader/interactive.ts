import { Key } from '#console/key';
import { dataTableNavigationAction, startsDataTableSearch } from '#console/output/data-table/keys';
import { cancelDataTableSelection } from '#console/output/data-table/reader/cancel';
import { assertSelectedDataTableRow, dataTableSelectionResult } from '#console/output/data-table/reader/result';
import { createDataTableReaderSession } from '#console/output/data-table/reader/session';
import type { DataTableReadOptions } from '#console/output/data-table/reader/types';
import type { DataTableSelectionReadResult } from '#console/output/data-table/types';
import type { PromptInput } from '#console/types';

export const readDataTableSelectionInteractive = async <T>(
	readKey: NonNullable<PromptInput['readKey']>,
	options: DataTableReadOptions<T>,
	headers: string[],
): Promise<DataTableSelectionReadResult<T>> => {
	const session = createDataTableReaderSession(options, headers);

	session.render();

	while (true) {
		const key = await readKey();

		if (key === null) {
			const currentRows = session.rows();

			assertSelectedDataTableRow(currentRows, session.selected(), session.frame());

			return dataTableSelectionResult(currentRows, session.selected(), false, false, session.frame());
		}

		const rows = session.rows();

		if (key === Key.ctrlC) {
			return cancelDataTableSelection(options, headers, rows, session.selected(), session.frame(), session.mode(), session.query());
		}

		if (session.applySearchKey(key)) {
			continue;
		}

		if (startsDataTableSearch(key)) {
			session.beginSearch();
			continue;
		}

		const action = dataTableNavigationAction(key);

		if (action !== null) {
			session.moveSelection(action);
			continue;
		}

		if (key === Key.enter) {
			assertSelectedDataTableRow(rows, session.selected(), session.frame());

			return dataTableSelectionResult(rows, session.selected(), true, false, session.frame());
		}
	}
};

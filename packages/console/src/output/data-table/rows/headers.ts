import { parseDataTableArrayHeaders, parseDataTableRecordHeaders, parseDataTableRowShape } from '#console/output/validators/data-table';
import { dataTableRowFields } from '#console/output/data-table/rows/fields';
import type { DataTableRow } from '#console/types';

export const deriveDataTableHeaders = <T>(rows: Array<DataTableRow<T>>): string[] => {
	const first = rows[0];

	if (!first) {
		return [];
	}

	const shape = parseDataTableRowShape(first);

	if (shape.kind === 'array') {
		return parseDataTableArrayHeaders(shape.row);
	}

	return parseDataTableRecordHeaders(dataTableRowFields(first));
};

import { tableCellArraySchema, tableCellRecordSchema } from '#console/output/validators/cells';
import type { TableCell, TableOptions } from '#console/types';

export const inferredTableHeaders = (rows: TableOptions['rows']): string[] => {
	const firstRow = rows[0];

	const arrayRow = tableCellArraySchema.safeParse(firstRow);

	if (arrayRow.success || firstRow === undefined) {
		return [];
	}

	return Object.keys(firstRow);
};

export const stringifyTableCell = (value: TableCell): string => {
	return value === null || value === undefined ? '' : String(value);
};

export const tableRowCells = (row: TableOptions['rows'][number], headers: string[]): string[] => {
	const arrayRow = tableCellArraySchema.safeParse(row);

	if (arrayRow.success) {
		return arrayRow.data.map(stringifyTableCell);
	}

	const recordRow = tableCellRecordSchema.safeParse(row);

	if (!recordRow.success) {
		throw new TypeError('Table rows must be arrays or records.');
	}

	return headers.map((header) => stringifyTableCell(recordRow.data[header]));
};

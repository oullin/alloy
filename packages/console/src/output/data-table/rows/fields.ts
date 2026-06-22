import { parseDataTableArrayRowFields, parseDataTableRowShape, parseDataTableRowValue } from '#console/output/validators/data-table';
import { stringifyDataTableCell } from '#console/output/data-table/rows/cells';
import type { DataTableRow, TableCell } from '#console/types';

export const dataTableRowFields = <T>(row: DataTableRow<T>): Record<string, TableCell> => {
	const shape = parseDataTableRowShape(row);

	if (shape.kind === 'array') {
		return parseDataTableArrayRowFields(shape.row);
	}

	if (shape.kind === 'object') {
		return shape.row.cells;
	}

	return shape.row;
};

export const dataTableRowCells = <T>(headers: string[], row: DataTableRow<T>): string[] => {
	const shape = parseDataTableRowShape(row);

	if (shape.kind === 'array') {
		return headers.length > 0 ? headers.map((_, index) => stringifyDataTableCell(shape.row[index])) : shape.row.map(stringifyDataTableCell);
	}

	const fields = dataTableRowFields(row);

	return headers.map((header) => stringifyDataTableCell(fields[header]));
};

export const dataTableRowValue = <T>(row: DataTableRow<T>, index: number): T | number => {
	const shape = parseDataTableRowShape(row);

	if (shape.kind === 'object' && shape.row.value !== undefined) {
		return shape.row.value;
	}

	if (shape.kind === 'record' && shape.row.value !== undefined) {
		return parseDataTableRowValue<T>(shape.row.value);
	}

	return index;
};

export const dataTableRowLabel = <T>(headers: string[], row: DataTableRow<T>): string => dataTableRowCells(headers, row).join(' ');

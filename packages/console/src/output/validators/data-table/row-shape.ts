import { z } from 'zod';
import { dataObjectRowSchema, dataTableArrayRowSchema, dataTableRecordRowSchema } from '#console/output/validators/data-table/schemas';
import type { DataTableObjectRow, DataTableRow, TableCell } from '#console/types';

const typedDataObjectRowSchema = <T>(): z.ZodType<DataTableObjectRow<T>> => dataObjectRowSchema as z.ZodType<DataTableObjectRow<T>>;
const dataTableRowValueSchema = <T>(): z.ZodType<T> => z.unknown() as z.ZodType<T>;

export type DataTableRowShape<T> =
	| {
			kind: 'array';
			row: TableCell[];
	  }
	| {
			kind: 'object';
			row: DataTableObjectRow<T>;
	  }
	| {
			kind: 'record';
			row: Record<string, TableCell>;
	  };

export const isDataObjectRow = <T>(row: DataTableRow<T>): row is DataTableObjectRow<T> => {
	return dataObjectRowSchema.safeParse(row).success;
};

export const parseDataTableRowShape = <T>(row: DataTableRow<T>): DataTableRowShape<T> => {
	const objectRow = typedDataObjectRowSchema<T>().safeParse(row);

	if (objectRow.success) {
		return { kind: 'object', row: objectRow.data };
	}

	const arrayRow = dataTableArrayRowSchema.safeParse(row);

	if (arrayRow.success) {
		return { kind: 'array', row: arrayRow.data };
	}

	const recordRow = dataTableRecordRowSchema.safeParse(row);

	if (!recordRow.success) {
		throw new TypeError('Data table rows must be arrays, records, or object rows.');
	}

	return { kind: 'record', row: recordRow.data };
};

export const parseDataTableRowValue = <T>(value: unknown): T => {
	const parsed = dataTableRowValueSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Data table row values must resolve to a typed value.');
	}

	return parsed.data;
};

export const parseDataTableArrayRowFields = (row: TableCell[]): Record<string, TableCell> => {
	return Object.fromEntries(row.map((value, index) => [String(index), value]));
};

export const parseDataTableArrayHeaders = (row: TableCell[]): string[] => {
	return row.map((_, index) => String(index + 1));
};

export const parseDataTableRecordHeaders = (row: Record<string, TableCell>): string[] => {
	return Object.keys(row).filter((key) => key !== 'value');
};

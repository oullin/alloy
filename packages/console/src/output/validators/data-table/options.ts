import { z } from 'zod';
import { dataTableHeadersSchema, dataTablePromptOptionsSchema, dataTableRowsSchema, dataTableStepNameSchema } from '#console/output/validators/data-table/schemas';
import type { DataTablePromptOptions, DataTableRow } from '#console/types';

const dataTableDefaultSchema = <T>(): z.ZodType<T | number> => z.unknown() as z.ZodType<T | number>;
const dataTablePromptOptionsTypedSchema = <T>(): z.ZodType<DataTablePromptOptions<T>> => dataTablePromptOptionsSchema as z.ZodType<DataTablePromptOptions<T>>;
const dataTableRowsTypedSchema = <T>(): z.ZodType<Array<DataTableRow<T>>> => dataTableRowsSchema as z.ZodType<Array<DataTableRow<T>>>;

export const isDataTablePromptOptions = <T>(value: unknown): value is DataTablePromptOptions<T> => {
	return dataTablePromptOptionsSchema.safeParse(value).success;
};

export const dataTableStepName = (value: unknown): string | undefined => {
	const result = dataTableStepNameSchema.safeParse(value);

	return result.success ? result.data : undefined;
};

export const parseDataTablePromptOptions = <T>(
	optionsOrHeaders: unknown = [],
	rows: unknown = null,
	scroll = 10,
	label = '',
	hint = '',
	required: DataTablePromptOptions<T>['required'] = false,
	validate: DataTablePromptOptions<T>['validate'] = undefined,
	transform: DataTablePromptOptions<T>['transform'] = undefined,
	filter: DataTablePromptOptions<T>['filter'] = undefined,
): DataTablePromptOptions<T> => {
	const promptOptions = dataTablePromptOptionsTypedSchema<T>().safeParse(optionsOrHeaders);

	if (promptOptions.success) {
		return promptOptions.data;
	}

	return {
		filter,
		headers: parseDataTableHeaders(optionsOrHeaders),
		hint,
		message: label,
		required,
		rows: parseDataTableRows<T>(rows ?? []),
		scroll,
		transform,
		validate,
	};
};

export const parseDataTableDefault = <T>(value: unknown): T | number => {
	const parsed = dataTableDefaultSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Data table defaults must resolve to a typed value or row number.');
	}

	return parsed.data;
};

export const parseInitialDataTableRow = <T>(value: unknown): DataTableRow<T> => {
	const rows = parseDataTableRows<T>(value);

	return rows[0];
};

const parseDataTableHeaders = (value: unknown): string[] => {
	const parsed = dataTableHeadersSchema.safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Data table headers must be an array of strings.');
	}

	return parsed.data;
};

const parseDataTableRows = <T>(value: unknown): Array<DataTableRow<T>> => {
	const parsed = dataTableRowsTypedSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Data table rows must be a non-empty array.');
	}

	return parsed.data;
};

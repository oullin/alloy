import { isNullOutputRows, parseOutputScroll, parseOutputStepName } from '#console/form/builder/validators/output/common';
import { dataTableStepName, isDataTablePromptOptions } from '#console/output/validators/data-table';
import { dataTableRowsSchema } from '#console/output/validators/data-table/schemas';
import type { z } from 'zod';
import type { DataTablePromptOptions, DataTableRow } from '#console/types';

const outputDataTableRowsSchema = <T>(): z.ZodType<Array<DataTableRow<T>>> => dataTableRowsSchema as z.ZodType<Array<DataTableRow<T>>>;

export type ResolvedDataTableFormArguments<T> = {
	name?: string;
	options: DataTablePromptOptions<T>;
};

export const parseOutputDataTableRows = <T>(value: unknown): Array<DataTableRow<T>> => {
	if (isNullOutputRows(value)) {
		return [];
	}

	const parsed = outputDataTableRowsSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Data table rows must be an array.');
	}

	return parsed.data;
};

export const resolveDataTableFormArguments = <T>(
	optionsOrHeaders: DataTablePromptOptions<T> | string[] = [],
	rowsOrName: Array<DataTableRow<T>> | null | string = null,
	scrollOrName: number | string = 10,
	label = '',
	hint = '',
	required: DataTablePromptOptions<T>['required'] = false,
	validate?: DataTablePromptOptions<T>['validate'],
	transform?: DataTablePromptOptions<T>['transform'],
	filter?: DataTablePromptOptions<T>['filter'],
	name?: string,
): ResolvedDataTableFormArguments<T> => {
	if (isDataTablePromptOptions<T>(optionsOrHeaders)) {
		return {
			name: dataTableStepName(rowsOrName),
			options: optionsOrHeaders,
		};
	}

	return {
		name: parseOutputStepName(scrollOrName) ?? name,
		options: {
			filter,
			headers: optionsOrHeaders,
			hint,
			message: label,
			required,
			rows: parseOutputDataTableRows<T>(rowsOrName),
			scroll: parseOutputScroll(scrollOrName, 10),
			transform,
			validate,
		},
	};
};

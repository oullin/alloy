import { isNullOutputRows } from '#console/form/builder/validators/output/common';
import { isTableOptions, tableStepName } from '#console/output/validators/table';
import { tableRowsSchema } from '#console/output/validators/table/schemas';
import type { TableOptions } from '#console/types';

export type ResolvedTableFormArguments =
	| {
			kind: 'options';
			name?: string;
			options: TableOptions;
	  }
	| {
			headers: string[];
			kind: 'rows';
			name?: string;
			rows: TableOptions['rows'] | null;
	  };

export const parseOutputTableRows = (value: unknown): TableOptions['rows'] | null => {
	if (isNullOutputRows(value)) {
		return null;
	}

	const parsed = tableRowsSchema.safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Table rows must be an array.');
	}

	return parsed.data;
};

export const resolveTableFormArguments = (headersOrOptions: TableOptions | string[] = [], rowsOrName: TableOptions['rows'] | string | null = null, name?: string): ResolvedTableFormArguments => {
	if (isTableOptions(headersOrOptions)) {
		return {
			kind: 'options',
			name: tableStepName(rowsOrName),
			options: headersOrOptions,
		};
	}

	const rowsStepName = tableStepName(rowsOrName);

	if (rowsStepName !== undefined) {
		return {
			headers: headersOrOptions,
			kind: 'rows',
			name: rowsStepName,
			rows: null,
		};
	}

	return {
		headers: headersOrOptions,
		kind: 'rows',
		name,
		rows: parseOutputTableRows(rowsOrName),
	};
};

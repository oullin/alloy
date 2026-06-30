import { tableHeadersSchema, tableOptionsSchema, tableRowsSchema, tableStepNameSchema } from '#console/output/validators/table/schemas';
import { inferredTableHeaders } from '#console/output/validators/table/rows';
import type { TableOptions } from '#console/types';

export { tableRowCells } from '#console/output/validators/table/rows';

export const isTableOptions = (value: unknown): value is TableOptions => {
	return tableOptionsSchema.safeParse(value).success;
};

export const tableStepName = (value: unknown): string | undefined => {
	const result = tableStepNameSchema.safeParse(value);

	return result.success ? result.data : undefined;
};

export const parseTableOptions = (headersOrOptions: unknown = [], rows: unknown = null): Required<TableOptions> => {
	const tableOptions = tableOptionsSchema.safeParse(headersOrOptions);

	if (tableOptions.success) {
		const headers = tableOptions.data.headers ?? inferredTableHeaders(tableOptions.data.rows);

		return { headers, rows: tableOptions.data.rows };
	}

	if (rows === null) {
		const rowOnlyOptions = tableRowsSchema.safeParse(headersOrOptions);

		if (rowOnlyOptions.success) {
			return { headers: [], rows: rowOnlyOptions.data };
		}
	}

	return {
		headers: parseTableHeaders(headersOrOptions),
		rows: rows === null ? [] : parseTableRows(rows),
	};
};

const parseTableHeaders = (value: unknown): string[] => {
	const parsed = tableHeadersSchema.safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Table headers must be an array of strings.');
	}

	return parsed.data;
};

const parseTableRows = (value: unknown): TableOptions['rows'] => {
	const parsed = tableRowsSchema.safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Table rows must be an array.');
	}

	return parsed.data;
};

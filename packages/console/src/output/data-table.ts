import { normalizeDataTablePromptOptions } from '#console/output/data-table/options';
import { runDataTablePrompt } from '#console/output/data-table/run';
import { parseDataTablePromptOptions } from '#console/output/validators/data-table';
import type { DataTablePromptOptions, DataTableRow } from '#console/types';

export function datatable<T = unknown>(options: DataTablePromptOptions<T>): Promise<T | number>;

export function datatable<T = unknown>(
	headers?: string[],
	rows?: Array<DataTableRow<T>> | null,
	scroll?: number,
	label?: string,
	hint?: string,
	required?: DataTablePromptOptions<T>['required'],
	validate?: DataTablePromptOptions<T>['validate'],
	transform?: DataTablePromptOptions<T>['transform'],
	filter?: DataTablePromptOptions<T>['filter'],
): Promise<T | number>;

export async function datatable<T = unknown>(
	optionsOrHeaders: DataTablePromptOptions<T> | string[] = [],
	rows: Array<DataTableRow<T>> | null = null,
	scroll = 10,
	label = '',
	hint = '',
	required: DataTablePromptOptions<T>['required'] = false,
	validate?: DataTablePromptOptions<T>['validate'],
	transform?: DataTablePromptOptions<T>['transform'],
	filter?: DataTablePromptOptions<T>['filter'],
): Promise<T | number> {
	const parsedOptions = parseDataTablePromptOptions<T>(optionsOrHeaders, rows, scroll, label, hint, required, validate, transform, filter);
	const options = normalizeDataTablePromptOptions(parsedOptions);

	return runDataTablePrompt(options);
}

import { dataTable, table } from '#console/output';
import { resolveTableFormArguments } from '#console/form/builder/validators/output';
import type { FormBuilder } from '#console/form/builder/index';
import type { TableOptions } from '#console/types';

export function dataTableFormStep(this: FormBuilder, options: TableOptions, name?: string): FormBuilder;

export function dataTableFormStep(this: FormBuilder, headers: string[], name: string): FormBuilder;

export function dataTableFormStep(this: FormBuilder, headersOrOptions?: TableOptions | string[], rows?: TableOptions['rows'] | null, name?: string): FormBuilder;

export function dataTableFormStep(this: FormBuilder, headersOrOptions: TableOptions | string[] = [], rowsOrName: TableOptions['rows'] | string | null = null, name?: string): FormBuilder {
	const resolved = resolveTableFormArguments(headersOrOptions, rowsOrName, name);

	if (resolved.kind === 'options') {
		return this.addSideEffect(() => dataTable(resolved.options), resolved.name);
	}

	return this.addSideEffect(() => dataTable(resolved.headers, resolved.rows), resolved.name);
}

export function tableFormStep(this: FormBuilder, options: TableOptions, name?: string): FormBuilder;

export function tableFormStep(this: FormBuilder, headers: string[], name: string): FormBuilder;

export function tableFormStep(this: FormBuilder, headersOrOptions?: TableOptions | string[], rows?: TableOptions['rows'] | null, name?: string): FormBuilder;

export function tableFormStep(this: FormBuilder, headersOrOptions: TableOptions | string[] = [], rowsOrName: TableOptions['rows'] | string | null = null, name?: string): FormBuilder {
	const resolved = resolveTableFormArguments(headersOrOptions, rowsOrName, name);

	if (resolved.kind === 'options') {
		return this.addSideEffect(() => table(resolved.options), resolved.name);
	}

	return this.addSideEffect(() => table(resolved.headers, resolved.rows), resolved.name);
}

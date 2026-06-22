import type { FormBuilder } from '#console/form/builder/index';
import type { DataTablePromptOptions, DataTableRow, TableOptions } from '#console/types';

export type OutputBuilderMethods = {
	alert(this: FormBuilder, message: string, name?: string): FormBuilder;
	clear(this: FormBuilder, name?: string): FormBuilder;
	dataTable(this: FormBuilder, options: TableOptions, name?: string): FormBuilder;
	dataTable(this: FormBuilder, headers: string[], name: string): FormBuilder;
	dataTable(this: FormBuilder, headersOrOptions?: TableOptions | string[], rows?: TableOptions['rows'] | null, name?: string): FormBuilder;
	datatable<T = unknown>(this: FormBuilder, options: DataTablePromptOptions<T>, name?: string): FormBuilder;
	datatable<T = unknown>(
		this: FormBuilder,
		headers?: string[],
		rows?: Array<DataTableRow<T>> | null,
		scroll?: number,
		label?: string,
		hint?: string,
		required?: DataTablePromptOptions<T>['required'],
		validate?: DataTablePromptOptions<T>['validate'],
		transform?: DataTablePromptOptions<T>['transform'],
		filter?: DataTablePromptOptions<T>['filter'],
		name?: string,
	): FormBuilder;
	error(this: FormBuilder, message: string, name?: string): FormBuilder;
	grid(this: FormBuilder, items?: Array<string | number | boolean>, maxWidth?: number, name?: string): FormBuilder;
	info(this: FormBuilder, message: string, name?: string): FormBuilder;
	intro(this: FormBuilder, message: string, name?: string): FormBuilder;
	note(this: FormBuilder, message: string, type?: string | null, name?: string): FormBuilder;
	notify(this: FormBuilder, title: string, body?: string, subtitle?: string, sound?: string, icon?: string, name?: string): FormBuilder;
	outro(this: FormBuilder, message: string, name?: string): FormBuilder;
	table(this: FormBuilder, options: TableOptions, name?: string): FormBuilder;
	table(this: FormBuilder, headers: string[], name: string): FormBuilder;
	table(this: FormBuilder, headersOrOptions?: TableOptions | string[], rows?: TableOptions['rows'] | null, name?: string): FormBuilder;
	title(this: FormBuilder, value: string, name?: string): FormBuilder;
	warning(this: FormBuilder, message: string, name?: string): FormBuilder;
};

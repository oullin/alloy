import { datatable } from '#console/output';
import { previousValue } from '#console/form/builder/previous';
import { resolveDataTableFormArguments } from '#console/form/builder/validators/output';
import type { FormBuilder } from '#console/form/builder/index';
import type { DataTablePromptOptions, DataTableRow } from '#console/types';

export function datatableFormStep<T = unknown>(this: FormBuilder, options: DataTablePromptOptions<T>, name?: string): FormBuilder;

export function datatableFormStep<T = unknown>(
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

export function datatableFormStep<T = unknown>(
	this: FormBuilder,
	optionsOrHeaders: DataTablePromptOptions<T> | string[] = [],
	rowsOrName: Array<DataTableRow<T>> | null | string = null,
	scrollOrName: number | string = 10,
	label = '',
	hint = '',
	required: DataTablePromptOptions<T>['required'] = false,
	validate: DataTablePromptOptions<T>['validate'] = undefined,
	transform: DataTablePromptOptions<T>['transform'] = undefined,
	filter: DataTablePromptOptions<T>['filter'] = undefined,
	name?: string,
): FormBuilder {
	const resolved = resolveDataTableFormArguments(optionsOrHeaders, rowsOrName, scrollOrName, label, hint, required, validate, transform, filter, name);

	return this.add(
		(_, previous) =>
			datatable<T>({
				...resolved.options,
				default: previousValue(previous, resolved.options.default),
			}),
		resolved.name,
	);
}

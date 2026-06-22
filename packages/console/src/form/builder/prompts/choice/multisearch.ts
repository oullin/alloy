import { multisearch } from '#console/prompts/choices';
import { previousArray } from '#console/form/builder/previous';
import { resolveMultiSearchFormArguments } from '#console/form/builder/prompts/validators/search';
import type { FormBuilder } from '#console/form/builder/index';
import type { MultiSearchPromptOptions } from '#console/types';

export function multisearchFormStep<T>(this: FormBuilder, options: MultiSearchPromptOptions<T>, name?: string): FormBuilder;

export function multisearchFormStep<T>(
	this: FormBuilder,
	label: string,
	options: MultiSearchPromptOptions<T>['options'],
	placeholder?: string,
	scroll?: number,
	required?: MultiSearchPromptOptions<T>['required'],
	validate?: MultiSearchPromptOptions<T>['validate'],
	hint?: string,
	name?: string,
	transform?: MultiSearchPromptOptions<T>['transform'],
	info?: MultiSearchPromptOptions<T>['info'],
): FormBuilder;

export function multisearchFormStep<T>(
	this: FormBuilder,
	optionsOrLabel: MultiSearchPromptOptions<T> | string,
	options?: MultiSearchPromptOptions<T>['options'] | string,
	placeholder = '',
	scroll = 5,
	required: MultiSearchPromptOptions<T>['required'] = false,
	validate: MultiSearchPromptOptions<T>['validate'] = undefined,
	hint = 'Use the space bar to select options.',
	name?: string,
	transform: MultiSearchPromptOptions<T>['transform'] = undefined,
	info: MultiSearchPromptOptions<T>['info'] = '',
): FormBuilder {
	const resolved = resolveMultiSearchFormArguments(optionsOrLabel, options, placeholder, scroll, required, validate, hint, name, transform, info);

	return this.add((_, previous) => multisearch<T>({ ...resolved.options, default: previousArray(previous, resolved.options.default ?? []) }), resolved.name);
}

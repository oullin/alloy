import { multiselect } from '#console/prompts/choices';
import { previousArray } from '#console/form/builder/previous';
import { resolveMultiSelectFormArguments } from '#console/form/builder/prompts/validators/select';
import type { FormBuilder } from '#console/form/builder/index';
import type { ChoiceOptions, MultiSelectPromptOptions } from '#console/types';

export function multiselectFormStep<T>(this: FormBuilder, options: MultiSelectPromptOptions<T>, name?: string): FormBuilder;

export function multiselectFormStep<T>(
	this: FormBuilder,
	label: string,
	options: ChoiceOptions<T>,
	defaultValue?: T[],
	scroll?: number,
	required?: boolean | string,
	validate?: MultiSelectPromptOptions<T>['validate'],
	hint?: string,
	name?: string,
	transform?: MultiSelectPromptOptions<T>['transform'],
	info?: MultiSelectPromptOptions<T>['info'],
): FormBuilder;

export function multiselectFormStep<T>(
	this: FormBuilder,
	optionsOrLabel: MultiSelectPromptOptions<T> | string,
	optionsOrName?: ChoiceOptions<T> | string,
	defaultValue: T[] = [],
	scroll = 5,
	required: boolean | string = false,
	validate: MultiSelectPromptOptions<T>['validate'] = undefined,
	hint = 'Use the space bar to select options.',
	name?: string,
	transform: MultiSelectPromptOptions<T>['transform'] = undefined,
	info: MultiSelectPromptOptions<T>['info'] = '',
): FormBuilder {
	const resolved = resolveMultiSelectFormArguments(optionsOrLabel, optionsOrName, defaultValue, scroll, required, validate, hint, name, transform, info);

	return this.add((_, previous) => multiselect<T>({ ...resolved.options, default: previousArray(previous, resolved.options.default ?? []) }), resolved.name);
}

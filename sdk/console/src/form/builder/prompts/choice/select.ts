import { select } from '#console/prompts/choices';
import { previousValue } from '#console/form/builder/previous';
import { resolveSelectFormArguments } from '#console/form/builder/prompts/validators/select';
import type { FormBuilder } from '#console/form/builder/index';
import type { ChoiceOptions, SelectPromptOptions } from '#console/types';

export function selectFormStep<T>(this: FormBuilder, options: SelectPromptOptions<T>, name?: string): FormBuilder;

export function selectFormStep<T>(
	this: FormBuilder,
	label: string,
	options: ChoiceOptions<T>,
	defaultValue?: T,
	scroll?: number,
	validate?: SelectPromptOptions<T>['validate'],
	hint?: string,
	required?: boolean | string,
	name?: string,
	transform?: SelectPromptOptions<T>['transform'],
	info?: SelectPromptOptions<T>['info'],
): FormBuilder;

export function selectFormStep<T>(
	this: FormBuilder,
	optionsOrLabel: SelectPromptOptions<T> | string,
	optionsOrName?: ChoiceOptions<T> | string,
	defaultValue?: T,
	scroll = 5,
	validate?: SelectPromptOptions<T>['validate'],
	hint = '',
	required: boolean | string = true,
	name?: string,
	transform?: SelectPromptOptions<T>['transform'],
	info: SelectPromptOptions<T>['info'] = '',
): FormBuilder {
	const resolved = resolveSelectFormArguments(optionsOrLabel, optionsOrName, defaultValue, scroll, validate, hint, required, name, transform, info);

	return this.add((_, previous) => select<T>({ ...resolved.options, default: previousValue(previous, resolved.options.default) }), resolved.name);
}

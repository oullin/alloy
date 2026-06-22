import { search } from '#console/prompts/choices';
import { previousValue } from '#console/form/builder/previous';
import { resolveSearchFormArguments } from '#console/form/builder/prompts/validators/search';
import type { FormBuilder } from '#console/form/builder/index';
import type { SearchPromptOptions } from '#console/types';

export function searchFormStep<T>(this: FormBuilder, options: SearchPromptOptions<T>, name?: string): FormBuilder;

export function searchFormStep<T>(
	this: FormBuilder,
	label: string,
	options: SearchPromptOptions<T>['options'],
	placeholder?: string,
	scroll?: number,
	validate?: SearchPromptOptions<T>['validate'],
	hint?: string,
	required?: SearchPromptOptions<T>['required'],
	name?: string,
	transform?: SearchPromptOptions<T>['transform'],
	info?: SearchPromptOptions<T>['info'],
): FormBuilder;

export function searchFormStep<T>(
	this: FormBuilder,
	optionsOrLabel: SearchPromptOptions<T> | string,
	options?: SearchPromptOptions<T>['options'] | string,
	placeholder = '',
	scroll = 5,
	validate: SearchPromptOptions<T>['validate'] = undefined,
	hint = '',
	required: SearchPromptOptions<T>['required'] = true,
	name?: string,
	transform: SearchPromptOptions<T>['transform'] = undefined,
	info: SearchPromptOptions<T>['info'] = '',
): FormBuilder {
	const resolved = resolveSearchFormArguments(optionsOrLabel, options, placeholder, scroll, validate, hint, required, name, transform, info);

	return this.add((_, previous) => search<T>({ ...resolved.options, default: previousValue(previous, resolved.options.default) }), resolved.name);
}

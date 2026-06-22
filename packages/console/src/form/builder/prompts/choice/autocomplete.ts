import { autocomplete } from '#console/prompts/choices';
import { previousString } from '#console/form/builder/previous';
import { resolveAutocompleteFormArguments } from '#console/form/builder/prompts/validators/suggest';
import type { FormBuilder } from '#console/form/builder/index';
import type { SuggestOptions } from '#console/prompts/choices';
import type { TextPromptOptions } from '#console/types';

export function autocompleteFormStep(this: FormBuilder, options: SuggestOptions, name?: string): FormBuilder;

export function autocompleteFormStep(
	this: FormBuilder,
	label: string,
	options: SuggestOptions['options'],
	placeholder?: string,
	defaultValue?: string,
	required?: boolean | string,
	validate?: TextPromptOptions['validate'],
	hint?: string,
	name?: string,
	transform?: TextPromptOptions['transform'],
	info?: SuggestOptions['info'],
): FormBuilder;

export function autocompleteFormStep(
	this: FormBuilder,
	optionsOrLabel: SuggestOptions | string,
	options?: SuggestOptions['options'] | string,
	placeholder = '',
	defaultValue = '',
	required: boolean | string = false,
	validate: TextPromptOptions['validate'] = undefined,
	hint = '',
	name?: string,
	transform: TextPromptOptions['transform'] = undefined,
	info: SuggestOptions['info'] = '',
): FormBuilder {
	const resolved = resolveAutocompleteFormArguments(optionsOrLabel, options, placeholder, defaultValue, required, validate, hint, name, transform, info);

	return this.add((_, previous) => autocomplete({ ...resolved.options, default: previousString(previous, resolved.options.default ?? '') }), resolved.name);
}

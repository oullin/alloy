import { text } from '#console/prompts/basic';
import { previousString } from '#console/form/builder/previous';
import { resolveTextFormArguments } from '#console/form/builder/prompts/validators/basic';
import type { FormBuilder } from '#console/form/builder/index';
import type { TextPromptOptions } from '#console/types';

export function textFormStep(this: FormBuilder, options: TextPromptOptions, name?: string): FormBuilder;

export function textFormStep(
	this: FormBuilder,
	label: string,
	placeholder?: string,
	defaultValue?: string,
	required?: boolean | string,
	validate?: TextPromptOptions['validate'],
	hint?: string,
	name?: string,
	transform?: TextPromptOptions['transform'],
): FormBuilder;

export function textFormStep(
	this: FormBuilder,
	optionsOrLabel: TextPromptOptions | string,
	placeholder = '',
	defaultValue = '',
	required: TextPromptOptions['required'] = false,
	validate: TextPromptOptions['validate'] = undefined,
	hint = '',
	name?: string,
	transform: TextPromptOptions['transform'] = undefined,
): FormBuilder {
	const resolved = resolveTextFormArguments(optionsOrLabel, placeholder, defaultValue, required, validate, hint, name, transform);

	return this.add((_, previous) => text({ ...resolved.options, default: previousString(previous, resolved.options.default ?? '') }), resolved.name);
}

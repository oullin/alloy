import { textarea } from '#console/prompts/basic';
import { previousString } from '#console/form/builder/previous';
import { resolveTextareaFormArguments } from '#console/form/builder/prompts/validators/basic';
import type { FormBuilder } from '#console/form/builder/index';
import type { TextareaPromptOptions } from '#console/types';

export function textareaFormStep(this: FormBuilder, options: TextareaPromptOptions, name?: string): FormBuilder;

export function textareaFormStep(
	this: FormBuilder,
	label: string,
	placeholder?: string,
	defaultValue?: string,
	required?: boolean | string,
	validate?: TextareaPromptOptions['validate'],
	hint?: string,
	rows?: number,
	name?: string,
	transform?: TextareaPromptOptions['transform'],
): FormBuilder;

export function textareaFormStep(
	this: FormBuilder,
	optionsOrLabel: TextareaPromptOptions | string,
	placeholder = '',
	defaultValue = '',
	required: TextareaPromptOptions['required'] = false,
	validate: TextareaPromptOptions['validate'] = undefined,
	hint = '',
	rows = 5,
	name?: string,
	transform: TextareaPromptOptions['transform'] = undefined,
): FormBuilder {
	const resolved = resolveTextareaFormArguments(optionsOrLabel, placeholder, defaultValue, required, validate, hint, rows, name, transform);

	return this.add((_, previous) => textarea({ ...resolved.options, default: previousString(previous, resolved.options.default ?? '') }), resolved.name);
}

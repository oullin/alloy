import { password } from '#console/prompts/basic';
import { previousString } from '#console/form/builder/previous';
import { resolvePasswordFormArguments } from '#console/form/builder/prompts/validators/basic';
import type { FormBuilder } from '#console/form/builder/index';
import type { TextPromptOptions } from '#console/types';

export function passwordFormStep(this: FormBuilder, options: TextPromptOptions, name?: string): FormBuilder;

export function passwordFormStep(
	this: FormBuilder,
	label: string,
	placeholder?: string,
	required?: boolean | string,
	validate?: TextPromptOptions['validate'],
	hint?: string,
	name?: string,
	transform?: TextPromptOptions['transform'],
): FormBuilder;

export function passwordFormStep(
	this: FormBuilder,
	optionsOrLabel: TextPromptOptions | string,
	placeholder = '',
	required: TextPromptOptions['required'] = false,
	validate: TextPromptOptions['validate'] = undefined,
	hint = '',
	name?: string,
	transform: TextPromptOptions['transform'] = undefined,
): FormBuilder {
	const resolved = resolvePasswordFormArguments(optionsOrLabel, placeholder, required, validate, hint, name, transform);

	return this.add((_, previous) => password({ ...resolved.options, default: previousString(previous, resolved.options.default ?? '') }), resolved.name);
}

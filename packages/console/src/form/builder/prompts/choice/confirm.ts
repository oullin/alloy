import { confirm } from '#console/prompts/choices';
import { previousBoolean } from '#console/form/builder/previous';
import { resolveConfirmFormArguments } from '#console/form/builder/prompts/validators/select';
import type { FormBuilder } from '#console/form/builder/index';
import type { ConfirmPromptOptions } from '#console/types';

export function confirmFormStep(this: FormBuilder, options: ConfirmPromptOptions, name?: string): FormBuilder;

export function confirmFormStep(
	this: FormBuilder,
	label: string,
	defaultValue?: boolean,
	yes?: string,
	no?: string,
	required?: boolean | string,
	validate?: ConfirmPromptOptions['validate'],
	hint?: string,
	name?: string,
	transform?: ConfirmPromptOptions['transform'],
): FormBuilder;

export function confirmFormStep(
	this: FormBuilder,
	optionsOrLabel: ConfirmPromptOptions | string,
	defaultValueOrName: boolean | string = true,
	yes = 'Yes',
	no = 'No',
	required: boolean | string = false,
	validate: ConfirmPromptOptions['validate'] = undefined,
	hint = '',
	name?: string,
	transform: ConfirmPromptOptions['transform'] = undefined,
): FormBuilder {
	const resolved = resolveConfirmFormArguments(optionsOrLabel, defaultValueOrName, yes, no, required, validate, hint, name, transform);

	return this.add((_, previous) => confirm({ ...resolved.options, default: previousBoolean(previous, resolved.options.default ?? true) }), resolved.name);
}

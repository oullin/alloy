import { labelNumberOptions } from '#console/form/builder/prompts/basic/number/options';
import { runLabelNumberFormStep, runObjectNumberFormStep } from '#console/form/builder/prompts/basic/number/step';
import { resolveNumberFormArguments } from '#console/form/builder/prompts/validators/basic';
import type { FormBuilder } from '#console/form/builder/index';
import type { NumberPromptOptions } from '#console/types';

export function numberFormStep(this: FormBuilder, options: NumberPromptOptions, name?: string): FormBuilder;

export function numberFormStep(
	this: FormBuilder,
	label: string,
	placeholder?: string,
	defaultValue?: number | string,
	required?: boolean | string,
	validate?: NumberPromptOptions['validate'],
	hint?: string,
	min?: number,
	max?: number,
	step?: number,
	name?: string,
	transform?: NumberPromptOptions['transform'],
): FormBuilder;

export function numberFormStep(
	this: FormBuilder,
	optionsOrLabel: NumberPromptOptions | string,
	placeholder = '',
	defaultValue: number | string = '',
	required: NumberPromptOptions['required'] = false,
	validate: NumberPromptOptions['validate'] = undefined,
	hint = '',
	min?: number,
	max?: number,
	step?: number,
	name?: string,
	transform: NumberPromptOptions['transform'] = undefined,
): FormBuilder {
	const resolved = resolveNumberFormArguments(optionsOrLabel, placeholder, defaultValue, required, validate, hint, min, max, step, name, transform, arguments.length);

	if (resolved.kind === 'options') {
		return this.add((_, previous) => runObjectNumberFormStep(resolved.options, previous), resolved.name);
	}

	return this.add((_, previous) => {
		const promptOptions = labelNumberOptions({
			defaultValue: resolved.defaultValue,
			hasLabelDefault: resolved.hasLabelDefault,
			hint: resolved.hint,
			label: resolved.label,
			max: resolved.max,
			min: resolved.min,
			placeholder: resolved.placeholder,
			previous,
			required: resolved.required,
			step: resolved.step,
			transform: resolved.transform,
			validate: resolved.validate,
		});

		return runLabelNumberFormStep(promptOptions);
	}, resolved.name);
}

import { normalizeNumberPromptOptions } from '#console/prompts/number/options';
import { runNumberPrompt } from '#console/prompts/number/run';
import type { NumberPromptOptions } from '#console/types';

export function number(options: NumberPromptOptions): Promise<number | string>;

export function number(
	label: string,
	placeholder?: string,
	defaultValue?: number | string,
	required?: boolean | string,
	validate?: NumberPromptOptions['validate'],
	hint?: string,
	min?: number,
	max?: number,
	step?: number,
	transform?: NumberPromptOptions['transform'],
): Promise<number | string>;

export async function number(
	message: string | NumberPromptOptions,
	placeholder = '',
	defaultValue: number | string = '',
	required: boolean | string = false,
	validate: NumberPromptOptions['validate'] = undefined,
	hint = '',
	min: number | undefined = undefined,
	max: number | undefined = undefined,
	step: number | undefined = undefined,
	transform: NumberPromptOptions['transform'] = undefined,
): Promise<number | string> {
	const options = normalizeNumberPromptOptions({
		defaultValue,
		hasDefaultArgument: arguments.length >= 3,
		hint,
		max,
		message,
		min,
		placeholder,
		required,
		step,
		transform,
		validate,
	});

	return runNumberPrompt(options);
}

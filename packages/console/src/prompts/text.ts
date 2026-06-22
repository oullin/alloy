import { normalizeTextPromptOptions } from '#console/prompts/text/options';
import { runTextPrompt } from '#console/prompts/text/run';
import type { TextPromptOptions } from '#console/types';

export function text(options: TextPromptOptions): Promise<string>;

export function text(
	label: string,
	placeholder?: string,
	defaultValue?: string,
	required?: boolean | string,
	validate?: TextPromptOptions['validate'],
	hint?: string,
	transform?: TextPromptOptions['transform'],
): Promise<string>;

export async function text(
	message: string | TextPromptOptions,
	placeholder = '',
	defaultValue = '',
	required: boolean | string = false,
	validate?: TextPromptOptions['validate'],
	hint = '',
	transform?: TextPromptOptions['transform'],
): Promise<string> {
	return runTextPrompt(
		normalizeTextPromptOptions({
			defaultValue,
			hint,
			message,
			placeholder,
			required,
			transform,
			validate,
		}),
	);
}

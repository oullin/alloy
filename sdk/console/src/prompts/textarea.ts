import { normalizeTextareaPromptOptions } from '#console/prompts/textarea/options';
import { runTextareaPrompt } from '#console/prompts/textarea/run';
import type { TextareaPromptOptions } from '#console/types';

/** Prompts for multi-line text in an inline editor. */
export function textarea(options: TextareaPromptOptions): Promise<string>;

export function textarea(
	label: string,
	placeholder?: string,
	defaultValue?: string,
	required?: boolean | string,
	validate?: TextareaPromptOptions['validate'],
	hint?: string,
	rows?: number,
	transform?: TextareaPromptOptions['transform'],
): Promise<string>;

export async function textarea(
	message: string | TextareaPromptOptions,
	placeholder = '',
	defaultValue = '',
	required: boolean | string = false,
	validate?: TextareaPromptOptions['validate'],
	hint = '',
	rows = 5,
	transform?: TextareaPromptOptions['transform'],
): Promise<string> {
	const options = normalizeTextareaPromptOptions({
		defaultValue,
		hint,
		message,
		placeholder,
		required,
		rows,
		transform,
		validate,
	});

	return runTextareaPrompt(options, rows);
}

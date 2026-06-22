import { normalizePasswordPromptOptions } from '#console/prompts/password/options';
import { runPasswordPrompt } from '#console/prompts/password/run';
import type { TextPromptOptions } from '#console/types';

export function password(options: TextPromptOptions): Promise<string>;

export function password(
	label: string,
	placeholder?: string,
	required?: boolean | string,
	validate?: TextPromptOptions['validate'],
	hint?: string,
	transform?: TextPromptOptions['transform'],
): Promise<string>;

export async function password(
	message: string | TextPromptOptions,
	placeholder = '',
	required: boolean | string = false,
	validate?: TextPromptOptions['validate'],
	hint = '',
	transform?: TextPromptOptions['transform'],
): Promise<string> {
	return runPasswordPrompt(
		normalizePasswordPromptOptions({
			hint,
			message,
			placeholder,
			required,
			transform,
			validate,
		}),
	);
}

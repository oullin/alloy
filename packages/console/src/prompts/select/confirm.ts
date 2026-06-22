import { confirmHasDefault, normalizeConfirmPromptOptions } from '#console/prompts/select/confirm-options';
import { runConfirmPrompt } from '#console/prompts/select/confirm-run';
import type { ConfirmPromptOptions } from '#console/types';

export function confirm(options: ConfirmPromptOptions): Promise<boolean>;

export function confirm(
	label: string,
	defaultValue?: boolean,
	yes?: string,
	no?: string,
	required?: boolean | string,
	validate?: ConfirmPromptOptions['validate'],
	hint?: string,
	transform?: ConfirmPromptOptions['transform'],
): Promise<boolean>;

export async function confirm(
	message: string | ConfirmPromptOptions,
	defaultValue = true,
	yes = 'Yes',
	no = 'No',
	required: boolean | string = false,
	validate: ConfirmPromptOptions['validate'] = undefined,
	hint = '',
	transform: ConfirmPromptOptions['transform'] = undefined,
): Promise<boolean> {
	return runConfirmPrompt(normalizeConfirmPromptOptions(message, defaultValue, yes, no, required, validate, hint, transform, confirmHasDefault(message, arguments.length, defaultValue)));
}

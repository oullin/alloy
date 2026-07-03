import { normalizeSelectPromptOptions, selectHasDefault } from '#console/prompts/select/options';
import { runSelectPrompt } from '#console/prompts/select/run';
import type { ChoiceOptions, SelectPromptOptions } from '#console/types';

/** Prompts the user to pick a single option from a list. */
export function select<T>(options: SelectPromptOptions<T>): Promise<T>;

export function select<T>(
	label: string,
	options: ChoiceOptions<T>,
	defaultValue?: T,
	scroll?: number,
	validate?: SelectPromptOptions<T>['validate'],
	hint?: string,
	required?: SelectPromptOptions<T>['required'],
	transform?: SelectPromptOptions<T>['transform'],
	info?: SelectPromptOptions<T>['info'],
): Promise<T>;

export async function select<T>(
	optionsOrLabel: SelectPromptOptions<T> | string,
	source?: ChoiceOptions<T>,
	defaultValue?: T,
	scroll = 5,
	validate?: SelectPromptOptions<T>['validate'],
	hint = '',
	required: SelectPromptOptions<T>['required'] = true,
	transform?: SelectPromptOptions<T>['transform'],
	info: SelectPromptOptions<T>['info'] = '',
): Promise<T> {
	return runSelectPrompt(
		normalizeSelectPromptOptions(optionsOrLabel, source, defaultValue, scroll, validate, hint, required, transform, info, selectHasDefault(optionsOrLabel, arguments.length, defaultValue)),
	);
}

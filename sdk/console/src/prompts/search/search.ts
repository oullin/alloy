import { normalizeSearchPromptOptions } from '#console/prompts/search/options';
import { runSearchPrompt } from '#console/prompts/search/run';
import type { ChoiceOptions, SearchPromptOptions } from '#console/types';

/** Prompts the user to filter choices by typing and pick a single match. */
export function search<T>(options: SearchPromptOptions<T>): Promise<T>;

export function search<T>(
	label: string,
	options: SearchPromptOptions<T>['options'],
	placeholder?: string,
	scroll?: number,
	validate?: SearchPromptOptions<T>['validate'],
	hint?: string,
	required?: SearchPromptOptions<T>['required'],
	transform?: SearchPromptOptions<T>['transform'],
	info?: SearchPromptOptions<T>['info'],
): Promise<T>;

export async function search<T>(
	optionsOrLabel: SearchPromptOptions<T> | string,
	source?: ChoiceOptions<T> | ((query: string) => Promise<ChoiceOptions<T>> | ChoiceOptions<T>),
	placeholder = '',
	scroll = 5,
	validate?: SearchPromptOptions<T>['validate'],
	hint = '',
	required: SearchPromptOptions<T>['required'] = true,
	transform?: SearchPromptOptions<T>['transform'],
	info: SearchPromptOptions<T>['info'] = '',
): Promise<T> {
	return runSearchPrompt(normalizeSearchPromptOptions(optionsOrLabel, source, placeholder, scroll, validate, hint, required, transform, info));
}

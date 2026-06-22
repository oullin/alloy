import { normalizeMultiSearchPromptOptions } from '#console/prompts/search/multi-options';
import { runMultiSearchPrompt } from '#console/prompts/search/multi-run';
import type { ChoiceOptions, MultiSearchPromptOptions } from '#console/types';

export function multisearch<T>(options: MultiSearchPromptOptions<T>): Promise<T[]>;

export function multisearch<T>(
	label: string,
	options: MultiSearchPromptOptions<T>['options'],
	placeholder?: string,
	scroll?: number,
	required?: MultiSearchPromptOptions<T>['required'],
	validate?: MultiSearchPromptOptions<T>['validate'],
	hint?: string,
	transform?: MultiSearchPromptOptions<T>['transform'],
	info?: MultiSearchPromptOptions<T>['info'],
): Promise<T[]>;

export async function multisearch<T>(
	optionsOrLabel: MultiSearchPromptOptions<T> | string,
	source?: ChoiceOptions<T> | ((query: string) => Promise<ChoiceOptions<T>> | ChoiceOptions<T>),
	placeholder = '',
	scroll = 5,
	required: MultiSearchPromptOptions<T>['required'] = false,
	validate: MultiSearchPromptOptions<T>['validate'] = undefined,
	hint = 'Use the space bar to select options.',
	transform: MultiSearchPromptOptions<T>['transform'] = undefined,
	info: MultiSearchPromptOptions<T>['info'] = '',
): Promise<T[]> {
	return runMultiSearchPrompt(normalizeMultiSearchPromptOptions(optionsOrLabel, source, placeholder, scroll, required, validate, hint, transform, info));
}

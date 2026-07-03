import { normalizeMultiSelectPromptOptions } from '#console/prompts/select/multiselect/options';
import { runMultiSelectPrompt } from '#console/prompts/select/multiselect/run';
import type { ChoiceOptions, MultiSelectPromptOptions } from '#console/types';

/** Prompts the user to pick any number of options from a list. */
export function multiselect<T>(options: MultiSelectPromptOptions<T>): Promise<T[]>;

export function multiselect<T>(
	label: string,
	options: ChoiceOptions<T>,
	defaultValue?: T[],
	scroll?: number,
	required?: MultiSelectPromptOptions<T>['required'],
	validate?: MultiSelectPromptOptions<T>['validate'],
	hint?: string,
	transform?: MultiSelectPromptOptions<T>['transform'],
	info?: MultiSelectPromptOptions<T>['info'],
): Promise<T[]>;

export async function multiselect<T>(
	optionsOrLabel: MultiSelectPromptOptions<T> | string,
	source?: ChoiceOptions<T>,
	defaultValue: T[] = [],
	scroll = 5,
	required: MultiSelectPromptOptions<T>['required'] = false,
	validate?: MultiSelectPromptOptions<T>['validate'],
	hint = 'Use the space bar to select options.',
	transform?: MultiSelectPromptOptions<T>['transform'],
	info: MultiSelectPromptOptions<T>['info'] = '',
): Promise<T[]> {
	return runMultiSelectPrompt(normalizeMultiSelectPromptOptions(optionsOrLabel, source, defaultValue, scroll, required, validate, hint, transform, info));
}

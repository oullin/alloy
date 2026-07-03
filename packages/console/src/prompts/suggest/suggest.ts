import { runTextSuggestionPrompt } from '#console/prompts/suggest/lifecycle';
import { readSuggestionValue } from '#console/prompts/suggest/read';
import { renderSubmittedSuggestion } from '#console/prompts/suggest/render';
import { suggestOptions } from '#console/prompts/suggest/options';
import type { SuggestOptions } from '#console/prompts/suggest/options';
import type { MaybePromise, TextPromptOptions } from '#console/types';

/** Prompts for text while showing a navigable suggestion list under the input. */
export function suggest(options: SuggestOptions): Promise<string>;

export function suggest(
	label: string,
	options: string[] | ((query: string) => MaybePromise<string[]>),
	placeholder?: string,
	defaultValue?: string,
	scroll?: number,
	required?: boolean | string,
	validate?: TextPromptOptions['validate'],
	hint?: string,
	transform?: TextPromptOptions['transform'],
	info?: SuggestOptions['info'],
): Promise<string>;

export async function suggest(
	message: string | SuggestOptions,
	source: string[] | ((query: string) => MaybePromise<string[]>) = [],
	placeholder = '',
	defaultValue = '',
	scroll = 5,
	required: boolean | string = false,
	validate?: TextPromptOptions['validate'],
	hint = '',
	transform?: TextPromptOptions['transform'],
	info: SuggestOptions['info'] = '',
): Promise<string> {
	const options = suggestOptions(message, source, placeholder, defaultValue, scroll, required, validate, hint, transform, info);

	return runTextSuggestionPrompt('suggest', options, readSuggestionValue, renderSubmittedSuggestion);
}

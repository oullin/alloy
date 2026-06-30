import type { FormBuilder } from '#console/form/builder/index';
import type { SuggestOptions } from '#console/prompts/choices';
import type { TextPromptOptions } from '#console/types';

export type SuggestPromptBuilderMethods = {
	suggest(this: FormBuilder, options: SuggestOptions, name?: string): FormBuilder;
	suggest(
		this: FormBuilder,
		label: string,
		options: SuggestOptions['options'],
		placeholder?: string,
		defaultValue?: string,
		scroll?: number,
		required?: boolean | string,
		validate?: TextPromptOptions['validate'],
		hint?: string,
		name?: string,
		transform?: TextPromptOptions['transform'],
		info?: SuggestOptions['info'],
	): FormBuilder;
};

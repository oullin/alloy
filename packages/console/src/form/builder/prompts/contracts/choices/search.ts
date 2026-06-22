import type { FormBuilder } from '#console/form/builder/index';
import type { SearchPromptOptions } from '#console/types';

export type SearchPromptBuilderMethods = {
	search<T>(this: FormBuilder, options: SearchPromptOptions<T>, name?: string): FormBuilder;
	search<T>(
		this: FormBuilder,
		label: string,
		options: SearchPromptOptions<T>['options'],
		placeholder?: string,
		scroll?: number,
		validate?: SearchPromptOptions<T>['validate'],
		hint?: string,
		required?: SearchPromptOptions<T>['required'],
		name?: string,
		transform?: SearchPromptOptions<T>['transform'],
		info?: SearchPromptOptions<T>['info'],
	): FormBuilder;
};

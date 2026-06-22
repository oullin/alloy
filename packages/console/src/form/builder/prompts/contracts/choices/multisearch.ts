import type { FormBuilder } from '#console/form/builder/index';
import type { MultiSearchPromptOptions } from '#console/types';

export type MultiSearchPromptBuilderMethods = {
	multisearch<T>(this: FormBuilder, options: MultiSearchPromptOptions<T>, name?: string): FormBuilder;
	multisearch<T>(
		this: FormBuilder,
		label: string,
		options: MultiSearchPromptOptions<T>['options'],
		placeholder?: string,
		scroll?: number,
		required?: MultiSearchPromptOptions<T>['required'],
		validate?: MultiSearchPromptOptions<T>['validate'],
		hint?: string,
		name?: string,
		transform?: MultiSearchPromptOptions<T>['transform'],
		info?: MultiSearchPromptOptions<T>['info'],
	): FormBuilder;
};

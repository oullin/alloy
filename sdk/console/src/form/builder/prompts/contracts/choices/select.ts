import type { FormBuilder } from '#console/form/builder/index';
import type { ChoiceOptions, MaybePromise, SelectPromptOptions } from '#console/types';

export type SelectPromptBuilderMethods = {
	select<T>(this: FormBuilder, options: SelectPromptOptions<T>, name?: string): FormBuilder;
	select<T>(
		this: FormBuilder,
		label: string,
		options: ChoiceOptions<T>,
		defaultValue?: T,
		scroll?: number,
		validate?: (value: T) => MaybePromise<string | null | undefined>,
		hint?: string,
		required?: boolean | string,
		name?: string,
		transform?: (value: T) => MaybePromise<T>,
		info?: SelectPromptOptions<T>['info'],
	): FormBuilder;
};

import type { FormBuilder } from '#console/form/builder/index';
import type { ChoiceOptions, MaybePromise, MultiSelectPromptOptions } from '#console/types';

export type MultiSelectPromptBuilderMethods = {
	multiselect<T>(this: FormBuilder, options: MultiSelectPromptOptions<T>, name?: string): FormBuilder;
	multiselect<T>(
		this: FormBuilder,
		label: string,
		options: ChoiceOptions<T>,
		defaultValue?: T[],
		scroll?: number,
		required?: boolean | string,
		validate?: (value: T[]) => MaybePromise<string | null | undefined>,
		hint?: string,
		name?: string,
		transform?: (value: T[]) => MaybePromise<T[]>,
		info?: MultiSelectPromptOptions<T>['info'],
	): FormBuilder;
};

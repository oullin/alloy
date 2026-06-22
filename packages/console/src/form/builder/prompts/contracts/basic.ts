import type { FormBuilder } from '#console/form/builder/index';
import type { NumberPromptOptions, TextareaPromptOptions, TextPromptOptions } from '#console/types';
import type { MaybePromise } from '#console/types';

export type BasicPromptBuilderMethods = {
	number(this: FormBuilder, options: NumberPromptOptions, name?: string): FormBuilder;
	number(
		this: FormBuilder,
		label: string,
		placeholder?: string,
		defaultValue?: number | string,
		required?: boolean | string,
		validate?: (value: number | string) => MaybePromise<string | null | undefined>,
		hint?: string,
		min?: number,
		max?: number,
		step?: number,
		name?: string,
		transform?: NumberPromptOptions['transform'],
	): FormBuilder;
	password(this: FormBuilder, options: TextPromptOptions, name?: string): FormBuilder;
	password(
		this: FormBuilder,
		label: string,
		placeholder?: string,
		required?: boolean | string,
		validate?: TextPromptOptions['validate'],
		hint?: string,
		name?: string,
		transform?: TextPromptOptions['transform'],
	): FormBuilder;
	text(this: FormBuilder, options: TextPromptOptions, name?: string): FormBuilder;
	text(
		this: FormBuilder,
		label: string,
		placeholder?: string,
		defaultValue?: string,
		required?: boolean | string,
		validate?: TextPromptOptions['validate'],
		hint?: string,
		name?: string,
		transform?: TextPromptOptions['transform'],
	): FormBuilder;
	textarea(this: FormBuilder, options: TextareaPromptOptions, name?: string): FormBuilder;
	textarea(
		this: FormBuilder,
		label: string,
		placeholder?: string,
		defaultValue?: string,
		required?: boolean | string,
		validate?: TextPromptOptions['validate'],
		hint?: string,
		rows?: number,
		name?: string,
		transform?: TextPromptOptions['transform'],
	): FormBuilder;
};

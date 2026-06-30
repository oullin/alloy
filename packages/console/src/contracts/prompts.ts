import type { MaybePromise } from '#console/contracts/base';
import type { ChoiceOptions, PromptInfo } from '#console/contracts/choices';
import type { BasePromptOptions } from '#console/contracts/options';

export type TextPromptOptions = BasePromptOptions<string> & {
	label?: string;
	options?: string[] | ((query: string) => MaybePromise<string[]>);
	placeholder?: string;
	transform?: (value: string) => MaybePromise<string>;
};

export type TextareaPromptOptions = TextPromptOptions & {
	rows?: number;
};

export type NumberPromptOptions = BasePromptOptions<number | string> & {
	label?: string;
	placeholder?: string;
	min?: number;
	max?: number;
	integer?: boolean;
	step?: number;
	transform?: (value: number | string) => MaybePromise<number | string>;
};

export type ConfirmPromptOptions = BasePromptOptions<boolean> & {
	label?: string;
	yes?: string;
	no?: string;
	transform?: (value: boolean) => MaybePromise<boolean>;
};

export type SelectPromptOptions<T> = BasePromptOptions<T> & {
	label?: string;
	options: ChoiceOptions<T>;
	scroll?: number;
	info?: PromptInfo<T>;
	transform?: (value: T) => MaybePromise<T>;
};

export type MultiSelectPromptOptions<T> = BasePromptOptions<T[]> & {
	label?: string;
	options: ChoiceOptions<T>;
	scroll?: number;
	info?: PromptInfo<T>;
	transform?: (value: T[]) => MaybePromise<T[]>;
};

export type SearchPromptOptions<T> = Omit<BasePromptOptions<T>, 'required'> & {
	label?: string;
	options: ChoiceOptions<T> | ((query: string) => MaybePromise<ChoiceOptions<T>>);
	placeholder?: string;
	required?: true | string;
	scroll?: number;
	info?: PromptInfo<T>;
	transform?: (value: T) => MaybePromise<T>;
};

export type MultiSearchPromptOptions<T> = BasePromptOptions<T[]> & {
	label?: string;
	options: ChoiceOptions<T> | ((query: string) => MaybePromise<ChoiceOptions<T>>);
	placeholder?: string;
	scroll?: number;
	info?: PromptInfo<T>;
	transform?: (value: T[]) => MaybePromise<T[]>;
};

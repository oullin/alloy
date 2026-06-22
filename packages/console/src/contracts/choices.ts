export type Choice<T = string> = {
	label: string;
	value: T;
	hint?: string;
	disabled?: boolean | string;
};

export type ChoiceInput<T = string> = Choice<T> | T;

export type ChoiceOptions<T = string> = Array<ChoiceInput<T>> | Record<string, string>;

export type PromptInfo<T> = string | ((value: T | null) => string | null | undefined);

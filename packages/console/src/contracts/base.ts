export type MaybePromise<T> = T | Promise<T>;

export type ValidationResult = string | null | undefined;

export type Validator<T> = (value: T) => MaybePromise<ValidationResult>;

export type PromptValue = string | number | boolean | string[] | number[];

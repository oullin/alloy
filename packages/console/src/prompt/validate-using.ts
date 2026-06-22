import { AsyncLocalStorage } from 'node:async_hooks';
import { parseGlobalValidationOptions } from '#console/prompt/validators/global-validation';
import type { BasePromptOptions, MaybePromise, ValidationResult } from '#console/types';

export type PromptGlobalValidator<T> = (value: T, options: BasePromptOptions<T>) => MaybePromise<ValidationResult>;

type PromptGlobalValidationState = {
	validator: PromptGlobalValidator<unknown> | undefined;
};

const scopedGlobalValidationState = new AsyncLocalStorage<PromptGlobalValidationState>();
const currentGlobalValidationState: PromptGlobalValidationState = { validator: undefined };

const promptGlobalValidationState = (): PromptGlobalValidationState => scopedGlobalValidationState.getStore() ?? currentGlobalValidationState;

export const validateUsing = (validator?: PromptGlobalValidator<unknown> | null): void => {
	promptGlobalValidationState().validator = validator ?? undefined;
};

export const globalValidationMessage = async <T>(value: T, options: BasePromptOptions<T>): Promise<ValidationResult> => {
	const validator = promptGlobalValidationState().validator;

	if (!validator) {
		return undefined;
	}

	return validator(value, parseGlobalValidationOptions(options));
};

export const withPromptGlobalValidationScope = <T>(callback: () => MaybePromise<T>): Promise<T> => {
	return Promise.resolve(scopedGlobalValidationState.run({ ...promptGlobalValidationState() }, callback));
};

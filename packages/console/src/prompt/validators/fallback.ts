import { functionSchema } from '#console/validators/function';
import type { z } from 'zod';
import type { PromptFallbackCondition } from '#console/prompt/fallback';
import type { PromptFallbackHandler } from '#console/prompt/fallback';

type PromptFallbackConditionCallback = Exclude<PromptFallbackCondition, boolean>;

const fallbackConditionCallbackSchema = functionSchema<PromptFallbackConditionCallback>();

const fallbackHandlerSchema = <TOptions, TResult>(): z.ZodType<PromptFallbackHandler<TOptions, TResult>> => functionSchema<PromptFallbackHandler<TOptions, TResult>>();

export const isFallbackConditionCallback = (condition: PromptFallbackCondition): condition is PromptFallbackConditionCallback => {
	return fallbackConditionCallbackSchema.safeParse(condition).success;
};

export const resolveFallbackCondition = (condition: PromptFallbackCondition): boolean => {
	return isFallbackConditionCallback(condition) ? condition() : condition;
};

export const parseFallbackHandler = <TOptions, TResult>(handler: unknown): PromptFallbackHandler<TOptions, TResult> => {
	const parsed = fallbackHandlerSchema<TOptions, TResult>().safeParse(handler);

	if (!parsed.success) {
		throw new TypeError('Prompt fallback handlers must be functions.');
	}

	return parsed.data;
};

import { z } from 'zod';
import { functionSchema } from '#console/validators/function';
import type { MaybePromise } from '#console/types';

const spinnerCallbackSchema = <T>(): z.ZodType<() => MaybePromise<T>> => functionSchema<() => MaybePromise<T>>();
const spinnerMessageSchema = z.string();

const spinnerOptionsSchema = z
	.object({
		message: spinnerMessageSchema.default(''),
	})
	.passthrough()
	.default({ message: '' });

export type ResolvedStatusOptions = z.output<typeof spinnerOptionsSchema>;

export const isSpinnerCallback = <T>(value: unknown): value is () => MaybePromise<T> => {
	return spinnerCallbackSchema<T>().safeParse(value).success;
};

export const isSpinnerMessage = (value: unknown): value is string => {
	return spinnerMessageSchema.safeParse(value).success;
};

export const parseSpinnerCallback = <T>(value: unknown): (() => MaybePromise<T>) => {
	const parsed = spinnerCallbackSchema<T>().safeParse(value);

	if (!parsed.success) {
		throw new Error('A spinner callback is required.');
	}

	return parsed.data;
};

export const parseSpinnerOptions = (value: unknown): ResolvedStatusOptions => {
	const parsed = spinnerOptionsSchema.safeParse(value);

	if (!parsed.success) {
		throw new Error('Spinner options must include a string message.');
	}

	return parsed.data;
};

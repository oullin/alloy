import { z } from 'zod';
import { functionSchema } from '#console/validators/function';

const iteratorMethodSchema = functionSchema<() => Iterator<unknown>>();
const asyncIteratorMethodSchema = functionSchema<() => AsyncIterator<unknown>>();
const stringValueSchema = z.string();

const hasIteratorMethod = (value: unknown): boolean => {
	const iterator = (value as Partial<Iterable<unknown>> | null | undefined)?.[Symbol.iterator];

	return iteratorMethodSchema.safeParse(iterator).success;
};

const hasAsyncIteratorMethod = (value: unknown): boolean => {
	const iterator = (value as Partial<AsyncIterable<unknown>> | null | undefined)?.[Symbol.asyncIterator];

	return asyncIteratorMethodSchema.safeParse(iterator).success;
};

export const iterableSchema = <T>(): z.ZodType<Iterable<T>> => z.custom<Iterable<T>>((value) => !stringValueSchema.safeParse(value).success && hasIteratorMethod(value));

export const asyncIterableSchema = <T>(): z.ZodType<AsyncIterable<T>> => z.custom<AsyncIterable<T>>((value) => hasAsyncIteratorMethod(value));

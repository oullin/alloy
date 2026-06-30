import { z } from 'zod';
import { iterableSchema } from '#console/validators/iterable';

const progressTotalArgumentSchema = z.union([z.number(), z.literal(Number.POSITIVE_INFINITY), z.literal(Number.NEGATIVE_INFINITY)]);
const progressMessageArgumentSchema = z.string();

const progressStepsArgumentSchema = <T>() => z.union([progressTotalArgumentSchema, iterableSchema<T>()]);

export const isProgressTotalArgument = (value: unknown): value is number => {
	return progressTotalArgumentSchema.safeParse(value).success;
};

export const progressMessageArgument = (value: unknown): string | undefined => {
	const parsed = progressMessageArgumentSchema.safeParse(value);

	return parsed.success ? parsed.data : undefined;
};

export const progressStepsArgument = <T>(value: Iterable<T> | number | string | undefined): Iterable<T> | number => {
	const steps = value ?? 0;
	const parsed = progressStepsArgumentSchema<T>().safeParse(steps);

	if (!parsed.success || progressMessageArgumentSchema.safeParse(steps).success) {
		throw new Error('Progress steps must be an iterable or a number.');
	}

	return parsed.data;
};

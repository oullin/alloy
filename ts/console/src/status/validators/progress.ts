import { z } from 'zod';
import { iterableSchema } from '#console/validators/iterable';

const progressNumberSchema = z.number().finite();
const progressTotalInputSchema = z.union([z.number(), z.literal(Number.POSITIVE_INFINITY), z.literal(Number.NEGATIVE_INFINITY)]);

type ProgressValuesInput<T> = { kind: 'steps'; steps: Iterable<T> } | { kind: 'total'; total: number };

export const parseProgressTotal = (total: unknown): number => {
	const parsed = progressNumberSchema.safeParse(total);
	const value = parsed.success ? Math.trunc(parsed.data) : 0;

	if (value <= 0) {
		throw new Error('Progress bar must have at least one item.');
	}

	return value;
};

export const parseProgressStep = (step: unknown): number => {
	const parsed = progressNumberSchema.safeParse(step);

	return parsed.success ? Math.trunc(parsed.data) : 0;
};

export const parseProgressValuesInput = <T>(steps: Iterable<T> | number): ProgressValuesInput<T> => {
	const total = progressTotalInputSchema.safeParse(steps);

	if (total.success) {
		return { kind: 'total', total: parseProgressTotal(total.data) };
	}

	const parsedSteps = iterableSchema<T>().safeParse(steps);

	if (!parsedSteps.success) {
		throw new Error('Progress steps must be an iterable or a number.');
	}

	return { kind: 'steps', steps: parsedSteps.data };
};

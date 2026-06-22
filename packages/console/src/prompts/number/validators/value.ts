import { z } from 'zod';
import type { NumberPromptOptions } from '#console/types';

export type NumberPromptValue = number | string;

export type NumberValidationResult = {
	error?: string;
	value?: NumberPromptValue;
};

const numericInputSchema = z.string();

const numericValueSchema = numericInputSchema.trim().min(1).transform(Number).pipe(z.number().finite());

export const parseNumericValue = (value: unknown): number | null => {
	const parsed = numericValueSchema.safeParse(value);

	return parsed.success ? parsed.data : null;
};

export const parseNumberInput = (input: unknown, options: Pick<NumberPromptOptions, 'integer' | 'max' | 'min'> = {}): NumberValidationResult => {
	const rawInput = numericInputSchema.safeParse(input);

	if (!rawInput.success) {
		return { error: 'Must be a number' };
	}

	const raw = rawInput.data;
	const normalized = raw.trim();

	if (normalized === '') {
		return { value: '' };
	}

	const parsed = numericValueSchema.safeParse(raw);

	if (!parsed.success) {
		return { error: 'Must be a number' };
	}

	const numeric = parsed.data;

	if (options.min !== undefined && numeric < options.min) {
		return { error: `Must be at least ${options.min}` };
	}

	if (options.max !== undefined && numeric > options.max) {
		return { error: `Must be at most ${options.max}` };
	}

	return { value: options.integer === false ? numeric : Math.trunc(numeric) };
};

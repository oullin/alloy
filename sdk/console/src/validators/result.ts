import { z } from 'zod';
import type { ValidationResult } from '#console/types';

const validationResultSchema = z.union([z.string(), z.null(), z.undefined()]);

export const parseValidationResult = (result: unknown): ValidationResult => {
	const parsed = validationResultSchema.safeParse(result);

	if (!parsed.success) {
		throw new Error('The validator must return a string or null.');
	}

	return parsed.data;
};

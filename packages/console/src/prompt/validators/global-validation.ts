import { z } from 'zod';
import type { BasePromptOptions } from '#console/types';

const globalValidationOptionsSchema = z.unknown() as z.ZodType<BasePromptOptions<unknown>>;

export const parseGlobalValidationOptions = (options: unknown): BasePromptOptions<unknown> => {
	const parsed = globalValidationOptionsSchema.safeParse(options);

	if (!parsed.success) {
		throw new TypeError('Global validation options must resolve to prompt options.');
	}

	return parsed.data;
};

import { z } from 'zod';
import type { SelectPromptOptions } from '#console/types';

const disallowedRequiredSchema = z.literal(false);

export const assertSelectOptions = <T>(options: SelectPromptOptions<T>): void => {
	if (disallowedRequiredSchema.safeParse(options.required).success) {
		throw new TypeError('Argument [required] must be true or a string.');
	}
};

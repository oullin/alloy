import { z } from 'zod';
import type { SearchPromptOptions } from '#console/types';

const disallowedRequiredSchema = z.literal(false);

export const assertSearchOptions = <T>(options: SearchPromptOptions<T>): void => {
	if (disallowedRequiredSchema.safeParse(options.required).success) {
		throw new TypeError('Argument [required] must be true or a string.');
	}
};

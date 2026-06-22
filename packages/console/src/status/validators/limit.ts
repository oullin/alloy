import { z } from 'zod';

const limitSchema = z.number().finite();

export const parseLogLimit = (limit: unknown, defaultValue: number): number => {
	const parsed = limitSchema.safeParse(limit);

	if (!parsed.success) {
		return defaultValue;
	}

	const value = Math.trunc(parsed.data);

	return value <= 0 ? defaultValue : value;
};

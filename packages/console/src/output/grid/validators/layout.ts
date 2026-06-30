import { z } from 'zod';

const gridWidthSchema = z.number().finite().positive();

export const parseGridWidth = (width: unknown, defaultValue: number): number => {
	const parsed = gridWidthSchema.safeParse(width);

	if (!parsed.success) {
		return defaultValue;
	}

	return Math.max(1, Math.trunc(parsed.data));
};

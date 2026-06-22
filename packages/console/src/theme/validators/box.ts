import { z } from 'zod';

const boxWidthSchema = z.number().finite().nonnegative();

export const parseBoxWidth = (width: unknown, defaultValue: number): number => {
	const parsed = boxWidthSchema.safeParse(width);

	if (!parsed.success) {
		return defaultValue;
	}

	return Math.trunc(parsed.data);
};

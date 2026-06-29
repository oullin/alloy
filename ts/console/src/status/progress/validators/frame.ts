import { z } from 'zod';

const progressFrameWidthSchema = z.number().finite();

export const progressFrameWidth = (width: unknown): number => {
	const parsed = progressFrameWidthSchema.safeParse(width);

	if (!parsed.success) {
		return 0;
	}

	return Math.max(0, Math.trunc(parsed.data));
};

import { z } from 'zod';

const stringWidthSchema = z.number().finite().positive();

export const parseStringWidth = (width: unknown): number | undefined => {
	const parsed = stringWidthSchema.safeParse(width);

	return parsed.success ? Math.floor(parsed.data) : undefined;
};

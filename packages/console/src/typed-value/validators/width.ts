import { z } from 'zod';

const typedValueWrapWidthSchema = z.number().finite().positive();

export const parseTypedValueWrapWidth = (width: unknown): number | undefined => {
	const parsed = typedValueWrapWidthSchema.safeParse(width);

	if (!parsed.success) {
		return undefined;
	}

	return Math.floor(parsed.data);
};

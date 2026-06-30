import { z } from 'zod';
import type { KeyboardEventLike } from '#console/key/types';

const keyboardEventSchema = z
	.object({
		ctrl: z.boolean().optional(),
		meta: z.boolean().optional(),
		name: z.string().optional(),
		sequence: z.string().optional(),
		shift: z.boolean().optional(),
	})
	.catch({});

export const parseKeyboardEvent = (event: unknown): KeyboardEventLike => {
	const parsed = keyboardEventSchema.safeParse(event);

	return parsed.success ? parsed.data : {};
};

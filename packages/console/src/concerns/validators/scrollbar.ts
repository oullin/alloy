import { z } from 'zod';

const scrollbarStateSchema = z
	.object({
		firstVisible: z.number().finite().nonnegative(),
		height: z.number().finite().positive(),
		total: z.number().finite().nonnegative(),
	})
	.transform(({ firstVisible, height, total }) => ({
		firstVisible: Math.floor(firstVisible),
		height: Math.floor(height),
		total: Math.floor(total),
	}));

export type ScrollbarState = z.infer<typeof scrollbarStateSchema>;

export const parseScrollbarState = (state: unknown): ScrollbarState | undefined => {
	const parsed = scrollbarStateSchema.safeParse(state);

	if (!parsed.success || parsed.data.total <= parsed.data.height) {
		return undefined;
	}

	return {
		...parsed.data,
		firstVisible: Math.min(parsed.data.firstVisible, parsed.data.total - parsed.data.height),
	};
};

import { z } from 'zod';

const numberStepSchema = z.number().finite().positive();

export const parseNumberStep = (step: unknown, integer?: boolean): number => {
	const parsed = numberStepSchema.safeParse(step);

	if (!parsed.success) {
		return 1;
	}

	const value = integer === false ? parsed.data : Math.trunc(parsed.data);

	return value > 0 ? value : 1;
};

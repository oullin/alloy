import { z } from 'zod';

const disabledReasonSchema = z.string();

export const parseDisabledReason = (value: unknown): string | undefined => {
	const parsed = disabledReasonSchema.safeParse(value);

	return parsed.success ? parsed.data : undefined;
};

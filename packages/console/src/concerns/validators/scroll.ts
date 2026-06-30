import { z } from 'zod';

const scrollSchema = z.number().finite().positive();

export const parseOptionalScrollSize = (scroll: unknown): number | undefined => {
	const parsed = scrollSchema.safeParse(scroll);

	return parsed.success ? Math.max(1, Math.trunc(parsed.data)) : undefined;
};

export const parseScrollSize = (scroll: unknown, defaultValue: number): number => {
	return parseOptionalScrollSize(scroll) ?? defaultValue;
};

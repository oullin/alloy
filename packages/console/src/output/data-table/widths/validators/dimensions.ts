import { z } from 'zod';

const dataTableWidthSchema = z.number().finite().positive();
const dataTableColumnCountSchema = z.number().finite().nonnegative();

export const parseDataTableColumnCount = (count: unknown): number => {
	const parsed = dataTableColumnCountSchema.safeParse(count);

	if (!parsed.success) {
		return 0;
	}

	return Math.floor(parsed.data);
};

export const parseDataTableWidth = (width: unknown, defaultValue: number): number => {
	const parsed = dataTableWidthSchema.safeParse(width);

	if (!parsed.success) {
		return defaultValue;
	}

	return Math.max(1, Math.trunc(parsed.data));
};

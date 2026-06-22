import { z } from 'zod';

const terminalDimensionSchema = z.number().finite().positive();
const terminalLineCountSchema = z.number().finite().nonnegative();

export const parseTerminalDimension = (value: unknown, defaultValue: number): number => {
	const parsed = terminalDimensionSchema.safeParse(value);

	if (!parsed.success) {
		return defaultValue;
	}

	return Math.trunc(parsed.data);
};

export const parseTerminalLineCount = (value: unknown): number => {
	const parsed = terminalLineCountSchema.safeParse(value);

	if (!parsed.success) {
		return 0;
	}

	return Math.floor(parsed.data);
};

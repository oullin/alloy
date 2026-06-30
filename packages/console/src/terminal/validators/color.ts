import { z } from 'zod';
import type { TerminalColor } from '#console/terminal/capabilities';

const terminalColorSchema = z.tuple([z.number().int().min(0).max(255), z.number().int().min(0).max(255), z.number().int().min(0).max(255)]);

export const parseTerminalColor = (value: unknown, fallback: TerminalColor): TerminalColor => {
	const parsed = terminalColorSchema.safeParse(value);

	return parsed.success ? parsed.data : fallback;
};

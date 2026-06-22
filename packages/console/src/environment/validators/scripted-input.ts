import { z } from 'zod';

const scriptedInputLinesSchema = z.array(z.string());

export const parseScriptedInputLines = (lines: unknown): string[] => {
	const parsed = scriptedInputLinesSchema.safeParse(lines);

	if (!parsed.success) {
		throw new Error('Scripted input must be an array of strings.');
	}

	return parsed.data;
};

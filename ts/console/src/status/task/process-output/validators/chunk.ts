import { z } from 'zod';

const processOutputChunkLinesSchema = z.string().transform((chunk) => chunk.split(/\r?\n/u));

export const parseProcessOutputChunkLines = (chunk: unknown): string[] => {
	const parsed = processOutputChunkLinesSchema.safeParse(chunk);

	if (!parsed.success) {
		throw new TypeError('Process output chunks must be strings.');
	}

	return parsed.data;
};

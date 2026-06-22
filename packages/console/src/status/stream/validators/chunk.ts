import { z } from 'zod';

const streamChunkSchema = z.string();
const streamChunkLinesSchema = streamChunkSchema.transform((chunk) => chunk.split(/\r?\n/u));

export const parseStreamChunk = (chunk: unknown): string => {
	const parsed = streamChunkSchema.safeParse(chunk);

	if (!parsed.success) {
		throw new TypeError('Stream chunks must be strings.');
	}

	return parsed.data;
};

export const parseStreamChunkLines = (chunk: unknown): string[] => {
	const parsed = streamChunkLinesSchema.safeParse(chunk);

	if (!parsed.success) {
		throw new TypeError('Stream chunks must be strings.');
	}

	return parsed.data;
};

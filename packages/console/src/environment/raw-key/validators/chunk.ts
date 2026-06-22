import { z } from 'zod';

const bufferChunkSchema = z.custom<Buffer>(Buffer.isBuffer);

const rawKeyChunkTextSchema = z.union([z.string(), bufferChunkSchema.transform((chunk) => chunk.toString('utf8'))]);

export const parseRawKeyChunkText = (chunk: unknown): string => {
	const parsed = rawKeyChunkTextSchema.safeParse(chunk);

	if (!parsed.success) {
		throw new Error('Raw key chunks must be strings or buffers.');
	}

	return parsed.data;
};

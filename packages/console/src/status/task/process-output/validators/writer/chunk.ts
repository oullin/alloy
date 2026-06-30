import { z } from 'zod';
import { parseProcessOutputCallback, parseProcessOutputEncoding } from '#console/status/task/process-output/validators/writer/callback';
import type { ProcessOutputCallback } from '#console/status/task/process-output/validators/writer/callback';

export type ResolvedProcessOutputWrite = {
	callback?: ProcessOutputCallback;
	content: string;
};

const processOutputStringSchema = z.string();

export const parseProcessOutputChunk = (chunk: string | Uint8Array, encoding?: BufferEncoding): string => {
	const parsed = processOutputStringSchema.safeParse(chunk);

	return parsed.success ? parsed.data : Buffer.from(chunk).toString(encoding);
};

export const resolveProcessOutputWrite = (
	chunk: string | Uint8Array,
	encodingOrCallback?: BufferEncoding | ProcessOutputCallback,
	callback?: ProcessOutputCallback,
): ResolvedProcessOutputWrite => {
	const encoding = parseProcessOutputEncoding(encodingOrCallback);

	return {
		callback: parseProcessOutputCallback(encodingOrCallback, callback),
		content: parseProcessOutputChunk(chunk, encoding),
	};
};

import { z } from 'zod';

const taskLogMessageLinesSchema = z.string().transform((message) => message.split(/\r?\n/u).filter((line) => line.length > 0));

export const parseTaskLogMessageLines = (message: unknown): string[] => {
	const parsed = taskLogMessageLinesSchema.safeParse(message);

	if (!parsed.success) {
		throw new TypeError('Task log messages must be strings.');
	}

	return parsed.data;
};

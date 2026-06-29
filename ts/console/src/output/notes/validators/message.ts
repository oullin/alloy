import { z } from 'zod';

const noteMessageLinesSchema = z.string().transform((message) => message.split(/\r?\n/u));

export const parseNoteMessageLines = (message: unknown): string[] => {
	const parsed = noteMessageLinesSchema.safeParse(message);

	if (!parsed.success) {
		throw new TypeError('Note messages must be strings.');
	}

	return parsed.data;
};

import { z } from 'zod';

const promptOutputContentSchema = z.string();

export const parsePromptOutputContent = (content: unknown): string => {
	const parsed = promptOutputContentSchema.safeParse(content);

	if (!parsed.success) {
		throw new Error('Prompt output content must be a string.');
	}

	return parsed.data;
};

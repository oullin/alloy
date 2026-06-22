import { z } from 'zod';

const isTerminalTitleCharacter = (character: string): boolean => {
	const code = character.codePointAt(0);

	return code === undefined || !((code >= 0 && code <= 31) || code === 127 || code === 155);
};

const terminalTitleSchema = z.string().transform((title) => Array.from(title).filter(isTerminalTitleCharacter).join(''));

export const parseTerminalTitle = (title: unknown): string => {
	const parsed = terminalTitleSchema.safeParse(title);

	if (!parsed.success) {
		throw new TypeError('Terminal titles must be strings.');
	}

	return parsed.data;
};

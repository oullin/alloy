import { isTextPromptLabel } from '#console/concerns/validators/text-options';
import type { TextPromptOptions } from '#console/types';

export const textOptions = (message: string | TextPromptOptions, fallback = ''): TextPromptOptions => {
	if (isTextPromptLabel(message)) {
		return { message, default: fallback };
	}

	return { ...message, default: message.default ?? fallback };
};

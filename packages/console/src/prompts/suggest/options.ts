import { isSuggestPromptLabel } from '#console/prompts/suggest/validators/overload';
import type { MaybePromise, TextPromptOptions } from '#console/types';

export type SuggestOptions = TextPromptOptions & {
	options: string[] | ((query: string) => MaybePromise<string[]>);
	scroll?: number;
	info?: string | ((value: string | null) => string | null | undefined);
};

export const suggestOptions = (
	message: string | SuggestOptions,
	options: string[] | ((query: string) => MaybePromise<string[]>) = [],
	placeholder = '',
	defaultValue = '',
	scroll = 5,
	required: boolean | string = false,
	validate?: TextPromptOptions['validate'],
	hint = '',
	transform?: TextPromptOptions['transform'],
	info: SuggestOptions['info'] = '',
): SuggestOptions => {
	if (isSuggestPromptLabel(message)) {
		return { message, label: message, options, placeholder, default: defaultValue, scroll, required, validate, hint, transform, info };
	}

	return { ...message, default: message.default ?? '' };
};

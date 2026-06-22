import { textOptions } from '#console/concerns/text-options';
import { isTextPromptLabel } from '#console/concerns/validators/text-options';
import type { TextPromptOptions } from '#console/types';

export type TextPromptArgumentOptions = {
	defaultValue: string;
	hint: string;
	message: string | TextPromptOptions;
	placeholder: string;
	required: boolean | string;
	transform?: TextPromptOptions['transform'];
	validate?: TextPromptOptions['validate'];
};

export const normalizeTextPromptOptions = (options: TextPromptArgumentOptions): TextPromptOptions => {
	if (!isTextPromptLabel(options.message)) {
		return textOptions(options.message);
	}

	return textOptions({
		default: options.defaultValue,
		hint: options.hint,
		label: options.message,
		message: options.message,
		placeholder: options.placeholder,
		required: options.required,
		transform: options.transform,
		validate: options.validate,
	});
};

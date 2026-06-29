import { isTextPromptLabel } from '#console/concerns/validators/text-options';
import type { TextareaPromptOptions } from '#console/types';

export type TextareaPromptArgumentOptions = {
	defaultValue: string;
	hint: string;
	message: string | TextareaPromptOptions;
	placeholder: string;
	required: boolean | string;
	rows: number;
	transform?: TextareaPromptOptions['transform'];
	validate?: TextareaPromptOptions['validate'];
};

export const normalizeTextareaPromptOptions = (options: TextareaPromptArgumentOptions): TextareaPromptOptions => {
	if (!isTextPromptLabel(options.message)) {
		return {
			...options.message,
			default: options.message.default ?? '',
			rows: options.message.rows ?? options.rows,
		};
	}

	return {
		default: options.defaultValue,
		hint: options.hint,
		label: options.message,
		message: options.message,
		placeholder: options.placeholder,
		required: options.required,
		rows: options.rows,
		transform: options.transform,
		validate: options.validate,
	};
};

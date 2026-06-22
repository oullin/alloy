import { hasNumberDefaultArgument, isNumberPromptLabel, parseNumberDefault } from '#console/prompts/number/validators/overload';
import { hasPromptDefault } from '#console/validators/default';
import type { NormalizedNumberPromptOptions } from '#console/prompts/number/defaults';
import type { NumberPromptOptions } from '#console/types';

export type NumberPromptArgumentOptions = {
	defaultValue: number | string;
	hasDefaultArgument: boolean;
	hint: string;
	max?: number;
	message: string | NumberPromptOptions;
	min?: number;
	placeholder: string;
	required: boolean | string;
	step?: number;
	transform?: NumberPromptOptions['transform'];
	validate?: NumberPromptOptions['validate'];
};

export const normalizeNumberPromptOptions = (options: NumberPromptArgumentOptions): NormalizedNumberPromptOptions => {
	const message = options.message;

	if (!isNumberPromptLabel(message)) {
		const hasDefault = hasPromptDefault(message);

		return {
			...message,
			default: hasDefault ? parseNumberDefault(message.default, '') : '',
			hasDefault,
		};
	}

	const hasDefault = hasNumberDefaultArgument(options.hasDefaultArgument, options.defaultValue);

	return {
		default: hasDefault ? options.defaultValue : '',
		hasDefault,
		hint: options.hint,
		label: message,
		max: options.max,
		message,
		min: options.min,
		placeholder: options.placeholder,
		required: options.required,
		step: options.step,
		transform: options.transform,
		validate: options.validate,
	};
};

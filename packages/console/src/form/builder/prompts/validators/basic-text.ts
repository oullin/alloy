import { isBasicPromptOptions, parseBasicStepName } from '#console/form/builder/prompts/validators/basic-label';
import type { TextPromptOptions } from '#console/types';

export type ResolvedTextFormArguments = {
	name?: string;
	options: TextPromptOptions;
};

export const resolveTextFormArguments = (
	optionsOrLabel: TextPromptOptions | string,
	placeholder = '',
	defaultValue = '',
	required: TextPromptOptions['required'] = false,
	validate: TextPromptOptions['validate'] = undefined,
	hint = '',
	name?: string,
	transform: TextPromptOptions['transform'] = undefined,
): ResolvedTextFormArguments => {
	if (isBasicPromptOptions(optionsOrLabel)) {
		return {
			name: parseBasicStepName(placeholder),
			options: optionsOrLabel,
		};
	}

	return {
		name,
		options: {
			message: optionsOrLabel,
			label: optionsOrLabel,
			placeholder,
			default: defaultValue,
			required,
			validate,
			hint,
			transform,
		},
	};
};

export const resolvePasswordFormArguments = (
	optionsOrLabel: TextPromptOptions | string,
	placeholder = '',
	required: TextPromptOptions['required'] = false,
	validate: TextPromptOptions['validate'] = undefined,
	hint = '',
	name?: string,
	transform: TextPromptOptions['transform'] = undefined,
): ResolvedTextFormArguments => {
	if (isBasicPromptOptions(optionsOrLabel)) {
		return {
			name: parseBasicStepName(placeholder),
			options: optionsOrLabel,
		};
	}

	return {
		name,
		options: {
			message: optionsOrLabel,
			label: optionsOrLabel,
			placeholder,
			default: '',
			required,
			validate,
			hint,
			transform,
		},
	};
};

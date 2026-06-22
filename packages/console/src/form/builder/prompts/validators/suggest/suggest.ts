import { isSuggestPromptOptions, parseSuggestSource, parseSuggestStepName } from '#console/form/builder/prompts/validators/suggest/common';
import type { SuggestOptions } from '#console/prompts/suggest/options';
import type { TextPromptOptions } from '#console/types';

export type ResolvedSuggestFormArguments = {
	name?: string;
	options: SuggestOptions;
};

export const resolveSuggestFormArguments = (
	optionsOrLabel: SuggestOptions | string,
	sourceOrName?: SuggestOptions['options'] | string,
	placeholder = '',
	defaultValue = '',
	scroll = 5,
	required: boolean | string = false,
	validate: TextPromptOptions['validate'] = undefined,
	hint = '',
	name?: string,
	transform: TextPromptOptions['transform'] = undefined,
	info: SuggestOptions['info'] = '',
): ResolvedSuggestFormArguments => {
	if (isSuggestPromptOptions(optionsOrLabel)) {
		return {
			name: parseSuggestStepName(sourceOrName),
			options: optionsOrLabel,
		};
	}

	return {
		name,
		options: {
			message: optionsOrLabel,
			label: optionsOrLabel,
			options: parseSuggestSource(sourceOrName),
			placeholder,
			default: defaultValue,
			scroll,
			required,
			validate,
			hint,
			transform,
			info,
		},
	};
};

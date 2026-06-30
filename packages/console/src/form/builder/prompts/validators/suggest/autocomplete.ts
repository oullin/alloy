import { isSuggestPromptOptions, parseSuggestSource, parseSuggestStepName } from '#console/form/builder/prompts/validators/suggest/common';
import type { SuggestOptions } from '#console/prompts/suggest/options';
import type { TextPromptOptions } from '#console/types';

export type ResolvedAutocompleteFormArguments = {
	name?: string;
	options: SuggestOptions;
};

export const resolveAutocompleteFormArguments = (
	optionsOrLabel: SuggestOptions | string,
	sourceOrName?: SuggestOptions['options'] | string,
	placeholder = '',
	defaultValue = '',
	required: boolean | string = false,
	validate?: TextPromptOptions['validate'],
	hint = '',
	name?: string,
	transform?: TextPromptOptions['transform'],
	info: SuggestOptions['info'] = '',
): ResolvedAutocompleteFormArguments => {
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
			required,
			validate,
			hint,
			transform,
			info,
		},
	};
};

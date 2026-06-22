import { isSearchPromptOptions, parseSearchChoiceSource, parseSearchStepName } from '#console/form/builder/prompts/validators/search/common';
import type { MultiSearchPromptOptions } from '#console/types';

export type ResolvedMultiSearchFormArguments<T> = {
	name?: string;
	options: MultiSearchPromptOptions<T>;
};

export const resolveMultiSearchFormArguments = <T>(
	optionsOrLabel: MultiSearchPromptOptions<T> | string,
	sourceOrName?: MultiSearchPromptOptions<T>['options'] | string,
	placeholder = '',
	scroll = 5,
	required: MultiSearchPromptOptions<T>['required'] = false,
	validate?: MultiSearchPromptOptions<T>['validate'],
	hint = 'Use the space bar to select options.',
	name?: string,
	transform?: MultiSearchPromptOptions<T>['transform'],
	info: MultiSearchPromptOptions<T>['info'] = '',
): ResolvedMultiSearchFormArguments<T> => {
	if (isSearchPromptOptions(optionsOrLabel)) {
		return {
			name: parseSearchStepName(sourceOrName),
			options: optionsOrLabel,
		};
	}

	return {
		name,
		options: {
			message: optionsOrLabel,
			label: optionsOrLabel,
			options: parseSearchChoiceSource<T>(sourceOrName),
			placeholder,
			scroll,
			required,
			validate,
			hint,
			transform,
			info,
		},
	};
};

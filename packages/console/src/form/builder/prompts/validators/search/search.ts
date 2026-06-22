import { isSearchPromptOptions, parseSearchChoiceSource, parseSearchStepName } from '#console/form/builder/prompts/validators/search/common';
import type { SearchPromptOptions } from '#console/types';

export type ResolvedSearchFormArguments<T> = {
	name?: string;
	options: SearchPromptOptions<T>;
};

export const resolveSearchFormArguments = <T>(
	optionsOrLabel: SearchPromptOptions<T> | string,
	sourceOrName?: SearchPromptOptions<T>['options'] | string,
	placeholder = '',
	scroll = 5,
	validate?: SearchPromptOptions<T>['validate'],
	hint = '',
	required: SearchPromptOptions<T>['required'] = true,
	name?: string,
	transform?: SearchPromptOptions<T>['transform'],
	info: SearchPromptOptions<T>['info'] = '',
): ResolvedSearchFormArguments<T> => {
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
			validate,
			hint,
			required,
			transform,
			info,
		},
	};
};

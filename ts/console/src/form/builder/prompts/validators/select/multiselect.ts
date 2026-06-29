import { isSelectPromptOptions, parseSelectChoiceOptions, parseSelectStepName } from '#console/form/builder/prompts/validators/select/common';
import type { ChoiceOptions, MultiSelectPromptOptions } from '#console/types';

export type ResolvedMultiSelectFormArguments<T> = {
	name?: string;
	options: MultiSelectPromptOptions<T>;
};

export const resolveMultiSelectFormArguments = <T>(
	optionsOrLabel: MultiSelectPromptOptions<T> | string,
	choicesOrName?: ChoiceOptions<T> | string,
	defaultValue: T[] = [],
	scroll = 5,
	required: boolean | string = false,
	validate?: MultiSelectPromptOptions<T>['validate'],
	hint = 'Use the space bar to select options.',
	name?: string,
	transform?: MultiSelectPromptOptions<T>['transform'],
	info: MultiSelectPromptOptions<T>['info'] = '',
): ResolvedMultiSelectFormArguments<T> => {
	if (isSelectPromptOptions(optionsOrLabel)) {
		return {
			name: parseSelectStepName(choicesOrName),
			options: optionsOrLabel,
		};
	}

	return {
		name,
		options: {
			message: optionsOrLabel,
			options: parseSelectChoiceOptions<T>(choicesOrName),
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

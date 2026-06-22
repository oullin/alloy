import { isSelectPromptOptions, parseSelectChoiceOptions, parseSelectStepName } from '#console/form/builder/prompts/validators/select/common';
import type { ChoiceOptions, SelectPromptOptions } from '#console/types';

export type ResolvedSelectFormArguments<T> = {
	name?: string;
	options: SelectPromptOptions<T>;
};

export const resolveSelectFormArguments = <T>(
	optionsOrLabel: SelectPromptOptions<T> | string,
	choicesOrName?: ChoiceOptions<T> | string,
	defaultValue?: T,
	scroll = 5,
	validate?: SelectPromptOptions<T>['validate'],
	hint = '',
	required: boolean | string = true,
	name?: string,
	transform?: SelectPromptOptions<T>['transform'],
	info: SelectPromptOptions<T>['info'] = '',
): ResolvedSelectFormArguments<T> => {
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
			validate,
			hint,
			required,
			transform,
			info,
		},
	};
};

import { hasSelectDefaultArgument, isSelectPromptLabel, parseSelectChoiceOptions } from '#console/prompts/select/validators/overload';
import { hasPromptDefault } from '#console/validators/default';
import type { NormalizedSelectPromptOptions } from '#console/prompts/select/defaults';
import type { ChoiceOptions, SelectPromptOptions } from '#console/types';

export const selectHasDefault = <T>(optionsOrLabel: SelectPromptOptions<T> | string, argumentCount: number, defaultValue?: T): boolean => {
	return isSelectPromptLabel(optionsOrLabel) ? hasSelectDefaultArgument(argumentCount, defaultValue) : hasPromptDefault(optionsOrLabel);
};

export const normalizeSelectPromptOptions = <T>(
	optionsOrLabel: SelectPromptOptions<T> | string,
	source: ChoiceOptions<T> | undefined,
	defaultValue: T | undefined,
	scroll: number,
	validate: SelectPromptOptions<T>['validate'],
	hint: string,
	required: SelectPromptOptions<T>['required'],
	transform: SelectPromptOptions<T>['transform'],
	info: SelectPromptOptions<T>['info'],
	hasDefault: boolean,
): NormalizedSelectPromptOptions<T> => {
	const isLabel = isSelectPromptLabel(optionsOrLabel);

	if (!isLabel) {
		return { ...optionsOrLabel, hasDefault };
	}

	return {
		default: defaultValue,
		hasDefault,
		hint,
		info,
		label: optionsOrLabel,
		message: optionsOrLabel,
		options: parseSelectChoiceOptions<T>(source),
		required,
		scroll,
		transform,
		validate,
	};
};

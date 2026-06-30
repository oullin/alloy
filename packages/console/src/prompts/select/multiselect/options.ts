import { isSelectPromptLabel, parseSelectChoiceOptions } from '#console/prompts/select/validators/overload';
import { parseMultiSelectDefault } from '#console/prompts/select/validators/defaults';
import type { ChoiceOptions, MultiSelectPromptOptions } from '#console/types';

export type NormalizedMultiSelectPromptOptions<T> = MultiSelectPromptOptions<T> & {
	default: T[];
};

export const normalizeMultiSelectPromptOptions = <T>(
	optionsOrLabel: MultiSelectPromptOptions<T> | string,
	source: ChoiceOptions<T> | undefined,
	defaultValue: T[],
	scroll: number,
	required: MultiSelectPromptOptions<T>['required'],
	validate: MultiSelectPromptOptions<T>['validate'],
	hint: string,
	transform: MultiSelectPromptOptions<T>['transform'],
	info: MultiSelectPromptOptions<T>['info'],
): NormalizedMultiSelectPromptOptions<T> => {
	const options = isSelectPromptLabel(optionsOrLabel)
		? { message: optionsOrLabel, label: optionsOrLabel, options: parseSelectChoiceOptions<T>(source), default: defaultValue, scroll, required, validate, hint, transform, info }
		: optionsOrLabel;

	return { ...options, default: parseMultiSelectDefault<T>(options.default ?? []) };
};

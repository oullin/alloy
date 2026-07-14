import { isSearchPromptLabel } from '#console/prompts/search/validators/overload';
import { parseMultiSearchDefault } from '#console/prompts/search/validators/defaults';
import { parseSearchChoiceSource } from '#console/prompts/search/validators/source';
import type { ChoiceOptions, MultiSearchPromptOptions } from '#console/types';

export type NormalizedMultiSearchPromptOptions<T> = MultiSearchPromptOptions<T> & {
	default: T[];
};

export const normalizeMultiSearchPromptOptions = <T>(
	optionsOrLabel: MultiSearchPromptOptions<T> | string,
	source: ChoiceOptions<T> | ((query: string) => Promise<ChoiceOptions<T>> | ChoiceOptions<T>) | undefined,
	placeholder: string,
	scroll: number,
	required: MultiSearchPromptOptions<T>['required'],
	validate: MultiSearchPromptOptions<T>['validate'],
	hint: string,
	transform: MultiSearchPromptOptions<T>['transform'],
	info: MultiSearchPromptOptions<T>['info'],
): NormalizedMultiSearchPromptOptions<T> => {
	const options: MultiSearchPromptOptions<T> = isSearchPromptLabel(optionsOrLabel)
		? { message: optionsOrLabel, label: optionsOrLabel, options: parseSearchChoiceSource<T>(source), placeholder, scroll, required, validate, hint, transform, info }
		: optionsOrLabel;

	return {
		...options,
		default: parseMultiSearchDefault<T>(options.default ?? []),
	};
};

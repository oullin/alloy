import { assertSearchOptions } from '#console/prompts/search/validators/options';
import { isSearchPromptLabel } from '#console/prompts/search/validators/overload';
import { parseSearchChoiceSource } from '#console/prompts/search/validators/source';
import { hasPromptDefault } from '#console/validators/default';
import type { NormalizedSearchPromptOptions } from '#console/prompts/search/defaults';
import type { ChoiceOptions, SearchPromptOptions } from '#console/types';

export const normalizeSearchPromptOptions = <T>(
	optionsOrLabel: SearchPromptOptions<T> | string,
	source: ChoiceOptions<T> | ((query: string) => Promise<ChoiceOptions<T>> | ChoiceOptions<T>) | undefined,
	placeholder: string,
	scroll: number,
	validate: SearchPromptOptions<T>['validate'],
	hint: string,
	required: SearchPromptOptions<T>['required'],
	transform: SearchPromptOptions<T>['transform'],
	info: SearchPromptOptions<T>['info'],
): NormalizedSearchPromptOptions<T> => {
	const isLabel = isSearchPromptLabel(optionsOrLabel);
	const hasDefault = isLabel ? false : hasPromptDefault(optionsOrLabel);

	const options: NormalizedSearchPromptOptions<T> = isLabel
		? { message: optionsOrLabel, label: optionsOrLabel, options: parseSearchChoiceSource<T>(source), hasDefault, placeholder, scroll, validate, hint, required, transform, info }
		: { ...optionsOrLabel, hasDefault, required: optionsOrLabel.required ?? true };

	assertSearchOptions(options);

	return options;
};

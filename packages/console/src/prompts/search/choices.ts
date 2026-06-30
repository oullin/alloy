import { normalizeSearchChoices } from '#console/concerns/choices';
import { resolveSearchChoiceSourceOptions } from '#console/prompts/search/validators/source';
import type { Choice, SearchPromptOptions } from '#console/types';

export const resolveSearchChoices = async <T>(source: SearchPromptOptions<T>['options'], query: string): Promise<Array<Choice<T>>> => {
	const options = await resolveSearchChoiceSourceOptions(source, query);

	return normalizeSearchChoices(options);
};

export const searchMessage = (message: string, query: string): string => {
	return query.length > 0 ? `${message} ${query}` : message;
};

export const lastEnabledIndex = <T>(choices: Array<Choice<T>>): number => {
	for (let index = choices.length - 1; index >= 0; index -= 1) {
		if (!choices[index]?.disabled) {
			return index;
		}
	}

	return 0;
};

import { ask } from '#console/prompt';
import { resolveLineMultiSearchChoices } from '#console/prompts/search/line-mode';
import { toggleSearchChoice } from '#console/prompts/search/selection';
import type { SearchSelection } from '#console/prompts/search/selection';
import type { Choice, MultiSearchPromptOptions } from '#console/types';

export type MultiSearchChoicesReadResult<T> = {
	cancelled: boolean;
	frame?: string;
	submitted: boolean;
	submittedLabels: string[];
	value: T[];
};

export const selectedSearchValues = <T>(selected: SearchSelection<T>): T[] => [...selected.keys()];

export const lineMultiSearchValues = async <T>(options: MultiSearchPromptOptions<T>): Promise<T[]> => {
	const query = (await ask(options.message, options.hint)).trim();

	return resolveLineMultiSearchChoices(options, query);
};

export const toggleHighlightedSearchChoice = <T>(selected: SearchSelection<T>, choices: Array<Choice<T>>, highlighted: number | null): void => {
	if (highlighted === null) {
		return;
	}

	const choice = choices[highlighted];

	if (choice) {
		toggleSearchChoice(selected, choice);
	}
};

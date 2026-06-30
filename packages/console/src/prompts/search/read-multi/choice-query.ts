import { applyTypedKey } from '#console/typed-value';
import { resolveSearchChoices } from '#console/prompts/search/choices';
import { displayedSearchChoices } from '#console/prompts/search/selection';
import type { SearchSelection } from '#console/prompts/search/selection';
import type { TypedValueState } from '#console/typed-value/types';
import type { Choice, MultiSearchPromptOptions } from '#console/types';

export type MultiSearchChoiceQuery<T> = {
	applyTypedInput(key: string): Promise<{ cancelled: boolean }>;
	displayedChoices(): Array<Choice<T>>;
	resolveChoices(): Promise<void>;
	value(): TypedValueState;
};

export const createMultiSearchChoiceQuery = <T>(options: MultiSearchPromptOptions<T>, selected: SearchSelection<T>, initialChoices: Array<Choice<T>>): MultiSearchChoiceQuery<T> => {
	let query: TypedValueState = { cursor: 0, value: '' };
	let choices: Array<Choice<T>> = initialChoices;

	let activeQueryValue = '';

	const resolveChoices = async (): Promise<void> => {
		const currentQuery = query.value;

		activeQueryValue = currentQuery;

		const resolved = await resolveSearchChoices(options.options, currentQuery);

		if (activeQueryValue === currentQuery) {
			choices = resolved;
		}
	};

	return {
		async applyTypedInput(key) {
			const next = applyTypedKey(query, key);

			if (next.cancelled) {
				return { cancelled: true };
			}

			query = { cursor: next.cursor, value: next.value };

			await resolveChoices();

			return { cancelled: false };
		},
		displayedChoices() {
			return displayedSearchChoices(choices, selected, query.value);
		},
		resolveChoices,
		value() {
			return query;
		},
	};
};

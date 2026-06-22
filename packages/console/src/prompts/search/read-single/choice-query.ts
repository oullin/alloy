import { applyTypedKey } from '#console/typed-value';
import { resolveSearchChoices } from '#console/prompts/search/choices';
import type { SearchReadOptions } from '#console/prompts/search/read-single/types';
import type { TypedValueState } from '#console/typed-value/types';
import type { Choice } from '#console/types';

export type SingleSearchChoiceQuery<T> = {
	applyTypedInput(key: string): Promise<{ cancelled: boolean }>;
	choices(): Array<Choice<T>>;
	resolveChoices(): Promise<void>;
	value(): TypedValueState;
};

export const createSingleSearchChoiceQuery = <T>(options: SearchReadOptions<T>, initialChoices: Array<Choice<T>>): SingleSearchChoiceQuery<T> => {
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
		choices() {
			return choices;
		},
		resolveChoices,
		value() {
			return query;
		},
	};
};

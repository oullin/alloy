import { firstEnabledIndex } from '#console/concerns/choices';
import { moveSelectHighlight } from '#console/prompts/select/keys';
import { markedChoiceIndexes, toggleAllEnabledChoices, toggleMarkedChoice } from '#console/prompts/select/multiple';
import type { SelectNavigationAction } from '#console/prompts/select/keys';
import type { Choice } from '#console/types';

export type MultipleChoicesSessionState = {
	marked(): Set<number>;
	move(action: SelectNavigationAction): void;
	selected(): number;
	toggleAll(): void;
	toggleIndex(index: number): boolean;
	toggleSelected(): void;
};

export const createMultipleChoicesSessionState = <T>(choices: Array<Choice<T>>, defaults: T[], scroll?: number): MultipleChoicesSessionState => {
	let selected = firstEnabledIndex(choices);
	let marked = markedChoiceIndexes(choices, defaults);

	return {
		marked() {
			return marked;
		},
		move(action) {
			selected = moveSelectHighlight(choices, selected, action, scroll);
		},
		selected() {
			return selected;
		},
		toggleAll() {
			marked = toggleAllEnabledChoices(choices, marked);
		},
		toggleIndex(index) {
			if (!choices[index] || choices[index]?.disabled) {
				return false;
			}

			marked = toggleMarkedChoice(choices, marked, index);

			return true;
		},
		toggleSelected() {
			marked = toggleMarkedChoice(choices, marked, selected);
		},
	};
};

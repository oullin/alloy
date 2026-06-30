import { createMultipleChoicesSessionFrame } from '#console/prompts/select/read-multiple/session/frame';
import { createMultipleChoicesSessionState } from '#console/prompts/select/read-multiple/session/state';
import type { SelectNavigationAction } from '#console/prompts/select/keys';
import type { Choice, MultiSelectPromptOptions } from '#console/types';

export type MultipleChoicesReaderSession = {
	frame(): string;
	marked(): Set<number>;
	move(action: SelectNavigationAction): void;
	render(): void;
	selected(): number;
	toggleAll(): void;
	toggleIndex(index: number): boolean;
	toggleSelected(): void;
};

export const createMultipleChoicesReaderSession = <T>(
	message: string,
	choices: Array<Choice<T>>,
	defaults: T[],
	scroll?: number,
	info?: MultiSelectPromptOptions<T>['info'],
): MultipleChoicesReaderSession => {
	const state = createMultipleChoicesSessionState(choices, defaults, scroll);
	const frame = createMultipleChoicesSessionFrame(message, choices, scroll, info);

	const render = (): void => frame.render(state.selected(), state.marked());

	return {
		frame() {
			return frame.value();
		},
		marked() {
			return state.marked();
		},
		move(action) {
			state.move(action);
			render();
		},
		render,
		selected() {
			return state.selected();
		},
		toggleAll() {
			state.toggleAll();
			render();
		},
		toggleIndex(index) {
			if (!state.toggleIndex(index)) {
				return false;
			}

			render();

			return true;
		},
		toggleSelected() {
			state.toggleSelected();
			render();
		},
	};
};

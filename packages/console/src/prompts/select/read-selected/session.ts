import { eraseRenderedFrame } from '#console/status/frame';
import { moveSelectHighlight } from '#console/prompts/select/keys';
import { defaultChoiceIndex } from '#console/prompts/select/read-selected/result';
import { renderSelectedChoice } from '#console/prompts/select/render';
import type { SelectNavigationAction } from '#console/prompts/select/keys';
import type { Choice, SelectPromptOptions } from '#console/types';

export type SelectedChoiceReaderSession = {
	frame(): string;
	move(action: SelectNavigationAction): void;
	render(): void;
	selected(): number;
};

export const createSelectedChoiceReaderSession = <T>(
	message: string,
	choices: Array<Choice<T>>,
	defaultValue: T | undefined,
	hasDefault: boolean,
	scroll?: number,
	info?: SelectPromptOptions<T>['info'],
): SelectedChoiceReaderSession => {
	let selected = defaultChoiceIndex(choices, defaultValue, hasDefault);
	let frame = '';

	function render(): void {
		if (frame.length > 0) {
			eraseRenderedFrame(frame);
		}

		frame = renderSelectedChoice(message, choices, selected, scroll, info);
	}

	return {
		frame() {
			return frame;
		},
		move(action) {
			selected = moveSelectHighlight(choices, selected, action, scroll);
			render();
		},
		render,
		selected() {
			return selected;
		},
	};
};

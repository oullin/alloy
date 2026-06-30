import { eraseRenderedFrame } from '#console/status/frame';
import { renderSearchChoices } from '#console/prompts/search/render';
import type { MultiSearchReaderState } from '#console/prompts/search/read-multi/state';
import type { MultiSearchPromptOptions } from '#console/types';

export type MultiSearchFrameRenderer = {
	current(): string;
	render(): void;
};

export const createMultiSearchFrameRenderer = <T>(options: MultiSearchPromptOptions<T>, state: MultiSearchReaderState<T>): MultiSearchFrameRenderer => {
	let frame = '';

	return {
		current() {
			return frame;
		},
		render() {
			if (frame.length > 0) {
				eraseRenderedFrame(frame);
			}

			const query = state.query();

			frame = renderSearchChoices(
				options.message,
				query.value,
				query.cursor,
				state.displayedChoices(),
				state.highlighted(),
				state.markedChoiceIndexes(),
				state.selectedLabels(),
				options.scroll,
				options.info,
				true,
				options.placeholder,
			);
		},
	};
};

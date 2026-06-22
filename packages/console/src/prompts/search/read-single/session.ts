import { createSearchReaderSessionFrame } from '#console/prompts/search/read-single/session/frame';
import { createSingleSearchReaderState } from '#console/prompts/search/read-single/state';
import type { SearchNavigationAction } from '#console/prompts/search/keys';
import type { SearchReaderSelection, SearchReadOptions } from '#console/prompts/search/read-single/types';
import type { TypedValueState } from '#console/typed-value/types';
import type { Choice } from '#console/types';

import { applySearchSessionTypedInput, clearSearchSessionHighlight, moveSearchSessionHighlight } from '#console/prompts/search/read-single/session/actions';

export type SearchReaderSession<T> = {
	applyTypedInput(key: string): Promise<{ cancelled: boolean }>;
	choices(): Array<Choice<T>>;
	clearHighlight(): void;
	defaultSelection(): Promise<SearchReaderSelection<T>>;
	frame(): string;
	highlighted(): number | null;
	move(action: SearchNavigationAction): Promise<void>;
	query(): TypedValueState;
	render(): void;
	selectedSelection(): SearchReaderSelection<T>;
};

export const createSearchReaderSession = async <T>(options: SearchReadOptions<T>, attempt: number): Promise<SearchReaderSession<T>> => {
	const state = await createSingleSearchReaderState(options, attempt);

	const frame = createSearchReaderSessionFrame(options, state);

	return {
		async applyTypedInput(key: string) {
			return applySearchSessionTypedInput(state, frame, key);
		},
		choices() {
			return state.choices();
		},
		clearHighlight() {
			clearSearchSessionHighlight(state, frame);
		},
		async defaultSelection() {
			return state.defaultSelection();
		},
		frame() {
			return frame.value();
		},
		highlighted() {
			return state.highlighted();
		},
		async move(action) {
			await moveSearchSessionHighlight(state, frame, action);
		},
		query() {
			return state.query();
		},
		render() {
			frame.render();
		},
		selectedSelection() {
			return state.selectedSelection();
		},
	};
};

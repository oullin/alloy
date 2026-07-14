import { createMultiSearchFrameRenderer } from '#console/prompts/search/read-multi/frame';
import { createMultiSearchReaderState } from '#console/prompts/search/read-multi/state';
import type { SearchNavigationAction } from '#console/prompts/search/keys';
import type { SearchSelection } from '#console/prompts/search/selection';
import type { TypedValueState } from '#console/typed-value/types';
import type { MultiSearchPromptOptions } from '#console/types';

import {
	applyMultiSearchSessionTypedInput,
	moveMultiSearchSessionHighlight,
	toggleAllMultiSearchSessionDisplayed,
	toggleMultiSearchSessionHighlighted,
} from '#console/prompts/search/read-multi/session/actions';

export type MultiSearchReaderSession<T> = {
	applyTypedInput(key: string): Promise<{ cancelled: boolean }>;
	frame(): string;
	highlighted(): number | null;
	move(action: SearchNavigationAction): Promise<void>;
	query(): TypedValueState;
	render(): void;
	selected(): SearchSelection<T>;
	selectedLabels(): string[];
	toggleAllDisplayed(): void;
	toggleHighlighted(): void;
};

export const createMultiSearchReaderSession = async <T>(options: MultiSearchPromptOptions<T>): Promise<MultiSearchReaderSession<T>> => {
	const state = await createMultiSearchReaderState(options);

	const frame = createMultiSearchFrameRenderer(options, state);

	return {
		async applyTypedInput(key: string) {
			return applyMultiSearchSessionTypedInput(state, frame, key);
		},
		frame() {
			return frame.current();
		},
		highlighted() {
			return state.highlighted();
		},
		async move(action) {
			await moveMultiSearchSessionHighlight(state, frame, action);
		},
		query() {
			return state.query();
		},
		render: () => frame.render(),
		selected() {
			return state.selected();
		},
		selectedLabels() {
			return state.selectedLabels();
		},
		toggleAllDisplayed() {
			toggleAllMultiSearchSessionDisplayed(state, frame);
		},
		toggleHighlighted() {
			toggleMultiSearchSessionHighlighted(state, frame);
		},
	};
};

import type { SearchNavigationAction } from '#console/prompts/search/keys';
import type { MultiSearchFrameRenderer } from '#console/prompts/search/read-multi/frame';
import type { MultiSearchReaderState } from '#console/prompts/search/read-multi/state';

export const applyMultiSearchSessionTypedInput = async <T>(state: MultiSearchReaderState<T>, frame: MultiSearchFrameRenderer, key: string): Promise<{ cancelled: boolean }> => {
	const next = await state.applyTypedInput(key);

	if (!next.cancelled) {
		frame.render();
	}

	return next;
};

export const moveMultiSearchSessionHighlight = async <T>(state: MultiSearchReaderState<T>, frame: MultiSearchFrameRenderer, action: SearchNavigationAction): Promise<void> => {
	await state.move(action);

	frame.render();
};

export const toggleAllMultiSearchSessionDisplayed = <T>(state: MultiSearchReaderState<T>, frame: MultiSearchFrameRenderer): void => {
	state.toggleAllDisplayed();
	frame.render();
};

export const toggleMultiSearchSessionHighlighted = <T>(state: MultiSearchReaderState<T>, frame: MultiSearchFrameRenderer): void => {
	state.toggleHighlighted();
	frame.render();
};

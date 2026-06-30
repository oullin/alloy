import type { SearchNavigationAction } from '#console/prompts/search/keys';
import type { SearchReaderSessionFrame } from '#console/prompts/search/read-single/session/frame';
import type { SingleSearchReaderState } from '#console/prompts/search/read-single/state';

export const applySearchSessionTypedInput = async <T>(state: SingleSearchReaderState<T>, frame: SearchReaderSessionFrame, key: string): Promise<{ cancelled: boolean }> => {
	const next = await state.applyTypedInput(key);

	if (!next.cancelled) {
		frame.render();
	}

	return next;
};

export const clearSearchSessionHighlight = <T>(state: SingleSearchReaderState<T>, frame: SearchReaderSessionFrame): void => {
	state.clearHighlight();
	frame.render();
};

export const moveSearchSessionHighlight = async <T>(state: SingleSearchReaderState<T>, frame: SearchReaderSessionFrame, action: SearchNavigationAction): Promise<void> => {
	await state.move(action);

	frame.render();
};

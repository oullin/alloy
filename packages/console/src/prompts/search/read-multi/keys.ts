import { Key } from '#console/key';
import { searchNavigationAction } from '#console/prompts/search/keys';
import type { MultiSearchReaderSession } from '#console/prompts/search/read-multi/session';

export const applyMultiSearchKey = async <T>(key: string, session: MultiSearchReaderSession<T>): Promise<boolean> => {
	const action = searchNavigationAction(key);

	if (action !== null && (action !== 'first' || session.highlighted() !== null) && (action !== 'last' || session.highlighted() !== null)) {
		await session.move(action);

		return true;
	}

	if (key === Key.ctrlE && session.highlighted() !== null) {
		return true;
	}

	if (key === Key.ctrlA && session.highlighted() !== null) {
		session.toggleAllDisplayed();

		return true;
	}

	if (key === Key.space && session.highlighted() !== null) {
		session.toggleHighlighted();

		return true;
	}

	return false;
};

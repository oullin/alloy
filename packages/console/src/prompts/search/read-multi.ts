import { promptEnvironment } from '#console/environment';
import { Key } from '#console/key';
import { cancelMultiSearchChoices } from '#console/prompts/search/read-multi/cancel';
import { applyMultiSearchKey } from '#console/prompts/search/read-multi/keys';
import { lineMultiSearchValues, selectedSearchValues } from '#console/prompts/search/read-multi/result';
import type { MultiSearchChoicesReadResult } from '#console/prompts/search/read-multi/result';
import { createMultiSearchReaderSession } from '#console/prompts/search/read-multi/session';
import type { MultiSearchReaderSession } from '#console/prompts/search/read-multi/session';
import type { MultiSearchPromptOptions } from '#console/types';

const multiSearchSelectionResult = <T>(session: MultiSearchReaderSession<T>, submitted: boolean): MultiSearchChoicesReadResult<T> => ({
	cancelled: false,
	frame: session.frame(),
	submitted,
	submittedLabels: session.selectedLabels(),
	value: selectedSearchValues(session.selected()),
});

export const readMultiSearchChoices = async <T>(options: MultiSearchPromptOptions<T>): Promise<MultiSearchChoicesReadResult<T>> => {
	const environment = promptEnvironment();

	if (!environment.input.readKey) {
		return { cancelled: false, submitted: false, submittedLabels: [], value: await lineMultiSearchValues(options) };
	}

	const session = await createMultiSearchReaderSession(options);

	session.render();

	while (true) {
		const key = await environment.input.readKey();

		if (key === null) {
			return multiSearchSelectionResult(session, true);
		}

		if (key === Key.enter) {
			return multiSearchSelectionResult(session, true);
		}

		if (key === Key.ctrlC) {
			return cancelMultiSearchChoices(options, session);
		}

		if (await applyMultiSearchKey(key, session)) {
			continue;
		}

		const next = await session.applyTypedInput(key);

		if (next.cancelled) {
			return cancelMultiSearchChoices(options, session);
		}
	}
};

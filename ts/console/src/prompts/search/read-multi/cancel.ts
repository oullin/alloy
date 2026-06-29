import { cancelPrompt } from '#console/prompt';
import { eraseRenderedFrame } from '#console/status/frame';
import { renderCancelledSearch } from '#console/prompts/search/render';
import { selectedSearchValues } from '#console/prompts/search/read-multi/result';
import type { MultiSearchChoicesReadResult } from '#console/prompts/search/read-multi/result';
import type { MultiSearchReaderSession } from '#console/prompts/search/read-multi/session';
import type { MultiSearchPromptOptions } from '#console/types';

export const cancelMultiSearchChoices = async <T>(options: MultiSearchPromptOptions<T>, session: MultiSearchReaderSession<T>): Promise<MultiSearchChoicesReadResult<T>> => {
	eraseRenderedFrame(session.frame());
	renderCancelledSearch(options.message, session.query().value, options.placeholder);

	const value = selectedSearchValues(session.selected());

	return {
		cancelled: true,
		submitted: false,
		submittedLabels: session.selectedLabels(),
		value: await cancelPrompt(value),
	};
};

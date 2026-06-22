import { cancelPrompt } from '#console/prompt';
import { renderCancelledSuggestion } from '#console/prompts/suggest/render';
import { suggestionReadResult } from '#console/prompts/suggest/read-result';
import { eraseRenderedFrame } from '#console/status/frame';
import type { SuggestOptions } from '#console/prompts/suggest/options';
import type { TextSuggestionReadResult } from '#console/prompts/suggest/read-result';
import type { SuggestReaderSession } from '#console/prompts/suggest/read/session';

export const cancelSuggestionValue = async (options: SuggestOptions, session: SuggestReaderSession): Promise<TextSuggestionReadResult> => {
	const value = session.state().value;

	eraseRenderedFrame(session.frame());
	renderCancelledSuggestion(options.message, value, options.placeholder);

	return suggestionReadResult(await cancelPrompt(value), true, true);
};

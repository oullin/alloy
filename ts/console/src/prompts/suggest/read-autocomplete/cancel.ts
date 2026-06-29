import { cancelPrompt } from '#console/prompt';
import { renderCancelledAutocomplete } from '#console/prompts/suggest/render-autocomplete';
import { suggestionReadResult } from '#console/prompts/suggest/read-result';
import { eraseRenderedFrame } from '#console/status/frame';
import type { AutocompleteReaderSession } from '#console/prompts/suggest/read-autocomplete/session';
import type { SuggestOptions } from '#console/prompts/suggest/options';
import type { TextSuggestionReadResult } from '#console/prompts/suggest/read-result';

export const cancelAutocompleteValue = async (options: SuggestOptions, session: AutocompleteReaderSession): Promise<TextSuggestionReadResult> => {
	const value = session.state().value;

	eraseRenderedFrame(session.frame());
	renderCancelledAutocomplete(options.message, value, options.placeholder);

	return suggestionReadResult(await cancelPrompt(value), true, true);
};

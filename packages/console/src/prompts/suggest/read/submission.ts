import { suggestionReadResult } from '#console/prompts/suggest/read-result';
import type { TextSuggestionReadResult } from '#console/prompts/suggest/read-result';
import type { SuggestReaderSession } from '#console/prompts/suggest/read/session';

export const currentSuggestionValue = (session: SuggestReaderSession): TextSuggestionReadResult => {
	return suggestionReadResult(session.state().value, true, false, session.frame());
};

export const highlightedSuggestionValue = (session: SuggestReaderSession): TextSuggestionReadResult => {
	const highlighted = session.highlighted();
	const match = highlighted === null ? undefined : session.matches()[highlighted];

	return suggestionReadResult(match ?? session.state().value, true, false, session.frame());
};

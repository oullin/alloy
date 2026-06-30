import { promptEnvironment } from '#console/environment';
import { Key } from '#console/key';
import { ask } from '#console/prompt';
import { clearsSuggestionHighlight, suggestNavigationAction } from '#console/prompts/suggest/keys';
import { cancelSuggestionValue } from '#console/prompts/suggest/read/cancel';
import { currentSuggestionValue, highlightedSuggestionValue } from '#console/prompts/suggest/read/submission';
import { suggestionReadResult } from '#console/prompts/suggest/read-result';
import { createSuggestReaderSession } from '#console/prompts/suggest/read/session';
import type { TextSuggestionReadResult } from '#console/prompts/suggest/read-result';
import type { SuggestOptions } from '#console/prompts/suggest/options';

export type SuggestReadResult = TextSuggestionReadResult;

export const readSuggestionValue = async (options: SuggestOptions): Promise<SuggestReadResult> => {
	const environment = promptEnvironment();

	if (!environment.input.readKey) {
		return suggestionReadResult(await ask(options.message, options.hint), false);
	}

	const session = await createSuggestReaderSession(options);

	session.render();

	while (true) {
		const key = await environment.input.readKey();

		if (key === null) {
			return highlightedSuggestionValue(session);
		}

		const action = suggestNavigationAction(key);

		if (action !== null && (action !== 'first' || session.highlighted() !== null) && (action !== 'last' || session.highlighted() !== null)) {
			await session.move(action);

			continue;
		}

		if (clearsSuggestionHighlight(key) && session.highlighted() !== null) {
			session.clearHighlight();
			continue;
		}

		if (key === Key.enter) {
			return highlightedSuggestionValue(session);
		}

		const next = await session.applyTypedInput(key);

		if (next.submitted) {
			return currentSuggestionValue(session);
		}

		if (next.cancelled) {
			return cancelSuggestionValue(options, session);
		}
	}
};

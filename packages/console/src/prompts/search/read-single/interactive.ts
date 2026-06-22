import { Key } from '#console/key';
import { clearsSearchHighlight, searchNavigationAction } from '#console/prompts/search/keys';
import { cancelInteractiveSearchChoice, cancelInteractiveSearchInput } from '#console/prompts/search/read-single/interactive/cancel';
import { defaultInteractiveSearchChoice, exhaustedInteractiveSearchChoice, highlightedInteractiveSearchChoice } from '#console/prompts/search/read-single/interactive/result';
import { createSearchReaderSession } from '#console/prompts/search/read-single/session';
import type { SearchChoiceReadResult } from '#console/prompts/search/read-single/result';
import type { SearchReadOptions } from '#console/prompts/search/read-single/types';
import type { PromptInput } from '#console/types';

export const readSearchChoiceInteractive = async <T>(readKey: NonNullable<PromptInput['readKey']>, options: SearchReadOptions<T>, attempt: number): Promise<SearchChoiceReadResult<T>> => {
	const session = await createSearchReaderSession(options, attempt);

	session.render();

	while (true) {
		const key = await readKey();

		if (key === null) {
			return exhaustedInteractiveSearchChoice(session, options);
		}

		if (key === Key.ctrlC) {
			return cancelInteractiveSearchChoice(session, options);
		}

		const action = searchNavigationAction(key, { controlNavigation: true, lineControls: true });

		if (action !== null && (action !== 'first' || session.highlighted() !== null) && (action !== 'last' || session.highlighted() !== null)) {
			await session.move(action);

			continue;
		}

		if (clearsSearchHighlight(key) && session.highlighted() !== null) {
			session.clearHighlight();
			continue;
		}

		if (key === Key.enter) {
			if (session.highlighted() !== null) {
				return highlightedInteractiveSearchChoice(session);
			}

			if (session.query().value === '' && options.hasDefault === true) {
				return defaultInteractiveSearchChoice(session);
			}

			session.clearHighlight();
			continue;
		}

		const next = await session.applyTypedInput(key);

		if (next.cancelled) {
			return cancelInteractiveSearchInput(session, options);
		}
	}
};

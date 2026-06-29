import { Key } from '#console/key';
import { selectNavigationAction } from '#console/prompts/select/keys';
import { parseChoiceIndex } from '#console/prompts/select/navigation';
import { cancelSelectedChoice } from '#console/prompts/select/read-selected/interactive/cancel';
import { exhaustedSelectedChoice, indexedSelectedChoice, submittedSelectedChoice } from '#console/prompts/select/read-selected/interactive/result';
import { createSelectedChoiceReaderSession } from '#console/prompts/select/read-selected/session';
import type { SelectedChoiceReadResult } from '#console/prompts/select/read-selected/types';
import type { Choice, PromptInput, SelectPromptOptions } from '#console/types';

export const readSelectedChoiceInteractive = async <T>(
	readKey: NonNullable<PromptInput['readKey']>,
	message: string,
	choices: Array<Choice<T>>,
	defaultValue?: T,
	hasDefault = false,
	scroll?: number,
	info?: SelectPromptOptions<T>['info'],
): Promise<SelectedChoiceReadResult<T>> => {
	const session = createSelectedChoiceReaderSession(message, choices, defaultValue, hasDefault, scroll, info);

	session.render();

	while (true) {
		const key = await readKey();

		if (key === null) {
			return exhaustedSelectedChoice(choices, session.selected());
		}

		if (key === Key.ctrlC) {
			return cancelSelectedChoice(message, choices, session, defaultValue, scroll);
		}

		const numeric = parseChoiceIndex(key);

		if (numeric !== null && choices[numeric - 1] && !choices[numeric - 1]?.disabled) {
			return indexedSelectedChoice(choices, numeric - 1, session.frame());
		}

		const action = selectNavigationAction(key, { lineControls: true });

		if (action !== null) {
			session.move(action);
			continue;
		}

		if (key === Key.enter) {
			return submittedSelectedChoice(choices, session);
		}
	}
};

import { Key } from '#console/key';
import { cancelMultipleChoices } from '#console/prompts/select/read-multiple/cancel';
import { applyMultipleChoicesKey } from '#console/prompts/select/read-multiple/keys';
import { multipleChoicesResult } from '#console/prompts/select/read-multiple/result';
import { createMultipleChoicesReaderSession } from '#console/prompts/select/read-multiple/session';
import type { MultipleChoicesReadResult } from '#console/prompts/select/read-multiple/types';
import type { Choice, MultiSelectPromptOptions } from '#console/types';

export const readMultipleChoicesInteractive = async <T>(
	readKey: () => Promise<string | null>,
	message: string,
	choices: Array<Choice<T>>,
	defaults: T[] = [],
	scroll?: number,
	info?: MultiSelectPromptOptions<T>['info'],
): Promise<MultipleChoicesReadResult<T>> => {
	const session = createMultipleChoicesReaderSession(message, choices, defaults, scroll, info);

	session.render();

	while (true) {
		const key = await readKey();

		if (key === null) {
			return multipleChoicesResult(choices, session.marked(), false, false, session.frame());
		}

		if (key === Key.ctrlC) {
			return cancelMultipleChoices(message, choices, session, scroll);
		}

		const applied = applyMultipleChoicesKey(key, choices, session);

		if (applied.handled) {
			if (applied.result) {
				return applied.result;
			}

			continue;
		}

		if (key === Key.enter) {
			return multipleChoicesResult(choices, session.marked(), true, false, session.frame());
		}
	}
};

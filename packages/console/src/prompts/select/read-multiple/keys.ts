import { Key } from '#console/key';
import { selectNavigationAction } from '#console/prompts/select/keys';
import { choicesFromCommaSeparated } from '#console/prompts/select/multiple';
import { parseChoiceIndex } from '#console/prompts/select/navigation';
import { multipleChoicesValueResult } from '#console/prompts/select/read-multiple/result';
import type { MultipleChoicesReaderSession } from '#console/prompts/select/read-multiple/session';
import type { MultipleChoicesReadResult } from '#console/prompts/select/read-multiple/types';
import type { Choice } from '#console/types';

export type MultipleChoicesKeyResult<T> =
	| {
			handled: false;
	  }
	| {
			handled: true;
			result?: MultipleChoicesReadResult<T>;
	  };

export const applyMultipleChoicesKey = <T>(key: string, choices: Array<Choice<T>>, session: MultipleChoicesReaderSession): MultipleChoicesKeyResult<T> => {
	if (key.includes(',')) {
		return { handled: true, result: multipleChoicesValueResult(choicesFromCommaSeparated(choices, key)) };
	}

	const numeric = parseChoiceIndex(key);

	if (numeric !== null && session.toggleIndex(numeric - 1)) {
		return { handled: true };
	}

	const action = selectNavigationAction(key);

	if (action !== null) {
		session.move(action);

		return { handled: true };
	}

	if (key === Key.ctrlA) {
		session.toggleAll();

		return { handled: true };
	}

	if (key === Key.space) {
		session.toggleSelected();

		return { handled: true };
	}

	return { handled: false };
};

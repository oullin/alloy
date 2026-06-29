import { selectedChoiceAt, selectedChoiceResult } from '#console/prompts/select/read-selected/result';
import type { SelectedChoiceReaderSession } from '#console/prompts/select/read-selected/session';
import type { SelectedChoiceReadResult } from '#console/prompts/select/read-selected/types';
import type { Choice } from '#console/types';

export const exhaustedSelectedChoice = <T>(choices: Array<Choice<T>>, selected: number): SelectedChoiceReadResult<T> => {
	const choice = selectedChoiceAt(choices, selected);

	return selectedChoiceResult(choice, false);
};

export const indexedSelectedChoice = <T>(choices: Array<Choice<T>>, index: number, frame: string): SelectedChoiceReadResult<T> => {
	const choice = selectedChoiceAt(choices, index);

	return selectedChoiceResult(choice, true, false, frame);
};

export const submittedSelectedChoice = <T>(choices: Array<Choice<T>>, session: SelectedChoiceReaderSession): SelectedChoiceReadResult<T> => {
	const choice = selectedChoiceAt(choices, session.selected());

	return selectedChoiceResult(choice, true, false, session.frame());
};

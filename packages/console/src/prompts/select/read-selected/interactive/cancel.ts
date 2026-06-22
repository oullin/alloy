import { cancelPrompt } from '#console/prompt';
import { parseChoiceValue } from '#console/concerns/validators/choice';
import { eraseRenderedFrame } from '#console/status/frame';
import { cancelledSelectedChoiceResult } from '#console/prompts/select/read-selected/result';
import { renderCancelledChoice } from '#console/prompts/select/render';
import type { SelectedChoiceReaderSession } from '#console/prompts/select/read-selected/session';
import type { SelectedChoiceReadResult } from '#console/prompts/select/read-selected/types';
import type { Choice } from '#console/types';

export const cancelSelectedChoice = async <T>(
	message: string,
	choices: Array<Choice<T>>,
	session: SelectedChoiceReaderSession,
	defaultValue?: T,
	scroll?: number,
): Promise<SelectedChoiceReadResult<T>> => {
	eraseRenderedFrame(session.frame());
	renderCancelledChoice(message, choices, session.selected(), scroll);

	const choice = choices[session.selected()];
	const fallback = choice?.value ?? defaultValue;

	const value = parseChoiceValue<T>(await cancelPrompt(fallback));

	return cancelledSelectedChoiceResult(choice, value);
};

import { cancelPrompt } from '#console/prompt';
import { eraseRenderedFrame } from '#console/status/frame';
import { markedChoiceValues } from '#console/prompts/select/multiple';
import { cancelledMultipleChoicesResult } from '#console/prompts/select/read-multiple/result';
import type { MultipleChoicesReaderSession } from '#console/prompts/select/read-multiple/session';
import type { MultipleChoicesReadResult } from '#console/prompts/select/read-multiple/types';
import { renderCancelledChoices } from '#console/prompts/select/render';
import type { Choice } from '#console/types';

export const cancelMultipleChoices = async <T>(message: string, choices: Array<Choice<T>>, session: MultipleChoicesReaderSession, scroll?: number): Promise<MultipleChoicesReadResult<T>> => {
	eraseRenderedFrame(session.frame());
	renderCancelledChoices(message, choices, session.selected(), session.marked(), scroll);

	const value = markedChoiceValues(choices, session.marked());

	return cancelledMultipleChoicesResult(choices, session.marked(), await cancelPrompt(value));
};

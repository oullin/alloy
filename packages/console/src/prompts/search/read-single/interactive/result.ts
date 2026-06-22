import type { SearchReaderSession } from '#console/prompts/search/read-single/session';
import type { SearchChoiceReadResult } from '#console/prompts/search/read-single/result';
import type { SearchReadOptions } from '#console/prompts/search/read-single/types';

export const fallbackSearchDefault = <T>(options: SearchReadOptions<T>): T | undefined => {
	return options.hasDefault === true ? options.default : undefined;
};

export const exhaustedInteractiveSearchChoice = async <T>(session: SearchReaderSession<T>, options: SearchReadOptions<T>): Promise<SearchChoiceReadResult<T>> => {
	if (session.highlighted() !== null) {
		return highlightedInteractiveSearchChoice(session);
	}

	if (session.query().value === '' && options.hasDefault === true) {
		const selected = await defaultInteractiveSearchChoice(session);

		if (selected.value !== undefined) {
			return selected;
		}
	}

	const choice = session.choices().find((candidate) => !candidate.disabled);

	return { cancelled: false, frame: session.frame(), submitted: choice !== undefined, submittedLabel: choice?.label ?? '', value: choice?.value ?? fallbackSearchDefault(options) };
};

export const highlightedInteractiveSearchChoice = <T>(session: SearchReaderSession<T>): SearchChoiceReadResult<T> => {
	const selected = session.selectedSelection();

	return { cancelled: false, frame: session.frame(), submitted: selected.submitted, submittedLabel: selected.label, value: selected.value };
};

export const defaultInteractiveSearchChoice = async <T>(session: SearchReaderSession<T>): Promise<SearchChoiceReadResult<T>> => {
	const selected = await session.defaultSelection();

	return { cancelled: false, frame: session.frame(), submitted: selected.submitted, submittedLabel: selected.label, value: selected.value };
};

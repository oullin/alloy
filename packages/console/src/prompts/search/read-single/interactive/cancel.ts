import { cancelPrompt } from '#console/prompt';
import { eraseRenderedFrame } from '#console/status/frame';
import { renderCancelledSearch } from '#console/prompts/search/render';
import { cancelledSearchValue } from '#console/prompts/search/read-single/result';
import { fallbackSearchDefault } from '#console/prompts/search/read-single/interactive/result';
import type { SearchReaderSession } from '#console/prompts/search/read-single/session';
import type { SearchChoiceReadResult } from '#console/prompts/search/read-single/result';
import type { SearchReadOptions } from '#console/prompts/search/read-single/types';

const cancelInteractiveSearch = async <T>(session: SearchReaderSession<T>, options: SearchReadOptions<T>, value: T | undefined): Promise<SearchChoiceReadResult<T>> => {
	eraseRenderedFrame(session.frame());
	renderCancelledSearch(options.message, session.query().value, options.placeholder);

	return { cancelled: true, submitted: false, submittedLabel: '', value: await cancelPrompt(value) };
};

export const cancelInteractiveSearchChoice = async <T>(session: SearchReaderSession<T>, options: SearchReadOptions<T>): Promise<SearchChoiceReadResult<T>> => {
	return cancelInteractiveSearch(session, options, cancelledSearchValue(session.choices(), session.highlighted(), options.default));
};

export const cancelInteractiveSearchInput = async <T>(session: SearchReaderSession<T>, options: SearchReadOptions<T>): Promise<SearchChoiceReadResult<T>> => {
	return cancelInteractiveSearch(session, options, fallbackSearchDefault(options));
};

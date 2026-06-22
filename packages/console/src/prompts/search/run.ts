import { promptUntilValid, promptWithFallback, PromptValidationError } from '#console/prompt';
import { activePromptFrame } from '#console/prompt/active-frame';
import { createPromptSubmissionState } from '#console/prompt/submission';
import { preserveSearchRetryDefault, transformSearchValue, transformedSearchDefault } from '#console/prompts/search/defaults';
import { readSearchChoice } from '#console/prompts/search/read-single';
import { renderSubmittedSearchChoice } from '#console/prompts/search/render';
import type { NormalizedSearchPromptOptions } from '#console/prompts/search/defaults';
import type { SearchPromptOptions } from '#console/types';

export const runSearchPrompt = async <T>(options: NormalizedSearchPromptOptions<T>): Promise<T> => {
	const validationOptions: SearchPromptOptions<T> = {
		...options,
		default: await transformedSearchDefault(options),
	};

	const activeFrame = activePromptFrame();
	const submission = createPromptSubmissionState('');

	return promptWithFallback('search', options, () =>
		promptUntilValid(
			validationOptions,
			async (attempt) => {
				submission.reset();

				const selected = await readSearchChoice(options, attempt);

				activeFrame.set(selected.frame);

				if (selected.value === undefined) {
					throw new PromptValidationError('Please select a valid option.');
				}

				submission.capture(selected.submitted, selected.cancelled, selected.submittedLabel);
				preserveSearchRetryDefault(options, selected.value);

				return transformSearchValue(options, selected.value);
			},
			() => {
				activeFrame.clear();
				submission.render((label) => {
					renderSubmittedSearchChoice(options.message, label);
				});
			},
			() => {
				activeFrame.clear();
				submission.reset();
			},
		),
	);
};

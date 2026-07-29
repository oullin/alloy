import { promptUntilValid, promptWithFallback } from '#console/prompt';
import { activePromptFrame } from '#console/prompt/active-frame';
import { createPromptSubmissionState } from '#console/prompt/submission';
import { preserveMultiSearchRetryDefault, transformedMultiSearchDefault } from '#console/prompts/search/defaults';
import { readMultiSearchChoices } from '#console/prompts/search/read-multi';
import { renderSubmittedSearchChoices } from '#console/prompts/search/render';
import type { NormalizedMultiSearchPromptOptions } from '#console/prompts/search/multi-options';
import type { MultiSearchPromptOptions } from '#console/types';

export const runMultiSearchPrompt = async <T>(options: NormalizedMultiSearchPromptOptions<T>): Promise<T[]> => {
	const validationOptions: MultiSearchPromptOptions<T> = {
		...options,
		default: await transformedMultiSearchDefault(options),
	};

	const activeFrame = activePromptFrame();

	const submission = createPromptSubmissionState<string[]>([]);

	return promptWithFallback('multisearch', options, () =>
		promptUntilValid(
			validationOptions,
			async () => {
				submission.reset();

				const selected = await readMultiSearchChoices(options);

				activeFrame.set(selected.frame);
				submission.capture(selected.submitted, selected.cancelled, selected.submittedLabels);
				preserveMultiSearchRetryDefault(options, selected.value);

				return options.transform ? options.transform(selected.value) : selected.value;
			},
			() => {
				activeFrame.clear();
				submission.render((labels) => {
					renderSubmittedSearchChoices(options.message, labels);
				});
			},
			() => {
				activeFrame.clear();
				submission.reset();
			},
		),
	);
};

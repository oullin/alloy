import { promptUntilValid, promptWithFallback } from '#console/prompt';
import { activePromptFrame } from '#console/prompt/active-frame';
import { createPromptSubmissionState } from '#console/prompt/submission';
import { preserveConfirmRetryDefault, transformedConfirmDefault, transformConfirmValue } from '#console/prompts/select/confirm-options';
import { readConfirm } from '#console/prompts/select/read-confirm';
import { renderSubmittedConfirm } from '#console/prompts/select/render-confirm';
import type { NormalizedConfirmPromptOptions } from '#console/prompts/select/confirm-options';
import type { ConfirmPromptOptions } from '#console/types';

export const runConfirmPrompt = async (options: NormalizedConfirmPromptOptions): Promise<boolean> => {
	const validationOptions: ConfirmPromptOptions = {
		...options,
		default: await transformedConfirmDefault(options),
	};

	const activeFrame = activePromptFrame();
	const submission = createPromptSubmissionState<void>(undefined);

	return promptWithFallback('confirm', options, () =>
		promptUntilValid(
			validationOptions,
			async () => {
				submission.reset();

				const answer = await readConfirm(options);

				activeFrame.set(answer.frame);
				submission.capture(answer.submitted, answer.cancelled, undefined);

				return transformConfirmValue(options, answer.value);
			},
			(value) => {
				activeFrame.clear();
				submission.render(() => {
					renderSubmittedConfirm(options, value);
				});
			},
			(value) => {
				preserveConfirmRetryDefault(options, value);
				activeFrame.clear();
				submission.reset();
			},
		),
	);
};

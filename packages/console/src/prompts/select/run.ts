import { normalizeChoices } from '#console/concerns/choices';
import { promptUntilValid, promptWithFallback } from '#console/prompt';
import { activePromptFrame } from '#console/prompt/active-frame';
import { createPromptSubmissionState } from '#console/prompt/submission';
import { preserveSelectRetryDefault, transformSelectValue, transformedSelectDefault } from '#console/prompts/select/defaults';
import { readSelectedChoice } from '#console/prompts/select/read-selected';
import { renderSubmittedChoice } from '#console/prompts/select/render';
import { assertSelectOptions } from '#console/prompts/select/validators/options';
import type { NormalizedSelectPromptOptions } from '#console/prompts/select/defaults';
import type { SelectPromptOptions } from '#console/types';

export const runSelectPrompt = async <T>(options: NormalizedSelectPromptOptions<T>): Promise<T> => {
	assertSelectOptions(options);

	const promptOptions: NormalizedSelectPromptOptions<T> = { ...options, required: options.required ?? true };

	const validationOptions: SelectPromptOptions<T> = {
		...promptOptions,
		default: await transformedSelectDefault(promptOptions),
	};

	const choices = normalizeChoices(options.options);
	const activeFrame = activePromptFrame();
	const submission = createPromptSubmissionState('');

	return promptWithFallback('select', promptOptions, () =>
		promptUntilValid(
			validationOptions,
			async () => {
				submission.reset();

				const selected = await readSelectedChoice(
					promptOptions.message,
					choices,
					promptOptions.default,
					promptOptions.hasDefault,
					promptOptions.hint,
					promptOptions.scroll,
					promptOptions.info,
				);

				activeFrame.set(selected.frame);
				submission.capture(selected.submitted, selected.cancelled, selected.submittedLabel);
				preserveSelectRetryDefault(promptOptions, selected.value);

				return transformSelectValue(promptOptions, selected.value);
			},
			() => {
				activeFrame.clear();
				submission.render((label) => {
					renderSubmittedChoice(promptOptions.message, label);
				});
			},
			() => {
				activeFrame.clear();
				submission.reset();
			},
		),
	);
};

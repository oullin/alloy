import { promptUntilValid, promptWithFallback } from '#console/prompt';
import { activePromptFrame } from '#console/prompt/active-frame';
import { createPromptSubmissionState } from '#console/prompt/submission';
import { transformedTextDefault } from '#console/prompts/text-default';
import { readTypedValue } from '#console/typed-value';
import { renderSubmittedTextareaFrame } from '#console/typed-value/textarea-frame';
import type { TextareaPromptOptions } from '#console/types';

export const runTextareaPrompt = async (options: TextareaPromptOptions, fallbackRows: number): Promise<string> => {
	const validationOptions: TextareaPromptOptions = {
		...options,
		default: await transformedTextDefault(options),
	};

	const activeFrame = activePromptFrame();
	const submission = createPromptSubmissionState<void>(undefined);

	return promptWithFallback('textarea', options, () =>
		promptUntilValid(
			validationOptions,
			async () => {
				submission.reset();

				const answer = await readTypedValue(options.message, {
					allowNewLine: true,
					default: options.default,
					hint: options.hint,
					placeholder: options.placeholder,
					rows: options.rows ?? fallbackRows,
				});

				const value = answer.value === '' && options.default !== undefined ? options.default : answer.value;

				activeFrame.set(answer.frame);
				submission.capture(!answer.cancelled, answer.cancelled, undefined);

				return options.transform ? options.transform(value) : value;
			},
			(value) => {
				activeFrame.clear();
				submission.render(() => {
					renderSubmittedTextareaFrame(options.message, value);
				});
			},
			(value) => {
				options.default = value;
				activeFrame.clear();
				submission.reset();
			},
		),
	);
};

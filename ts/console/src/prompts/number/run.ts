import { promptUntilValid, promptWithFallback, PromptValidationError } from '#console/prompt';
import { activePromptFrame } from '#console/prompt/active-frame';
import { createPromptSubmissionState } from '#console/prompt/submission';
import { numberDefault, preserveNumberRetryDefault, transformNumberValue } from '#console/prompts/number/defaults';
import { readNumberValue } from '#console/prompts/number/input';
import { renderSubmittedNumberValue } from '#console/prompts/number/render';
import { parseNumberInput } from '#console/prompts/number/validators/value';
import type { NormalizedNumberPromptOptions } from '#console/prompts/number/defaults';
import type { NumberPromptOptions } from '#console/types';

export const runNumberPrompt = async (options: NormalizedNumberPromptOptions): Promise<number | string> => {
	const validationOptions: NumberPromptOptions = {
		...options,
		default: await numberDefault(options),
	};

	const activeFrame = activePromptFrame();
	const submission = createPromptSubmissionState<void>(undefined);

	return promptWithFallback('number', options, () =>
		promptUntilValid(
			validationOptions,
			async () => {
				submission.reset();

				const answer = await readNumberValue(options.message, {
					default: options.default,
					hasDefault: options.hasDefault,
					hint: options.hint,
					integer: options.integer,
					max: options.max,
					min: options.min,
					placeholder: options.placeholder,
					step: options.step,
				});

				const value = answer.value;

				activeFrame.set(answer.frame);
				submission.capture(!answer.cancelled, answer.cancelled, undefined);

				const result = parseNumberInput(value, options);

				if (result.error !== undefined) {
					throw new PromptValidationError(result.error, value);
				}

				if (result.value === '' && options.hasDefault) {
					return numberDefault(options);
				}

				const parsedValue = result.value ?? '';

				return transformNumberValue(options, parsedValue);
			},
			(value) => {
				activeFrame.clear();
				submission.render(() => {
					renderSubmittedNumberValue(options.message, value);
				});
			},
			(value) => {
				if (value === undefined) {
					preserveNumberRetryDefault(options, value);
				} else {
					options.default = '';
					options.hasDefault = false;
				}

				activeFrame.clear();
				submission.reset();
			},
		),
	);
};

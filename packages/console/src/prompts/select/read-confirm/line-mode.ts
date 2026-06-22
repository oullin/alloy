import { ask } from '#console/prompt';
import type { ConfirmReadOptions, ConfirmReadResult } from '#console/prompts/select/read-confirm/types';

const confirmQuestionSuffix = (options: ConfirmReadOptions): string => {
	return options.hasDefault === true && options.default === false ? ' [y/N]' : ' [Y/n]';
};

export const readLineConfirm = async (options: ConfirmReadOptions): Promise<ConfirmReadResult> => {
	const answer = (await ask(`${options.message}${confirmQuestionSuffix(options)}`, options.hint)).trim().toLowerCase();

	if (answer === '') {
		return { cancelled: false, submitted: false, value: options.default ?? true };
	}

	return { cancelled: false, submitted: false, value: ['y', 'yes', options.yes?.toLowerCase()].includes(answer) };
};

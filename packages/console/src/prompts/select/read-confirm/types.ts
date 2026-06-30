import type { ConfirmPromptOptions } from '#console/types';

export type ConfirmReadOptions = ConfirmPromptOptions & {
	hasDefault?: boolean;
};

export type ConfirmReadResult = {
	cancelled: boolean;
	frame?: string;
	submitted: boolean;
	value: boolean;
};

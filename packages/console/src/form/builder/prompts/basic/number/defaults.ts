import { previousNumber } from '#console/form/builder/previous';
import { hasPromptDefault } from '#console/validators/default';
import { hasPreviousResponse } from '#console/form/builder/validators/previous';
import type { NumberPromptOptions } from '#console/types';

export const numberOptionsWithPreviousDefault = (options: NumberPromptOptions, previous: unknown): NumberPromptOptions => {
	if (hasPreviousResponse(previous)) {
		return { ...options, default: previousNumber(previous, options.default ?? '') };
	}

	if (hasPromptDefault(options)) {
		return { ...options, default: previousNumber(previous, options.default ?? '') };
	}

	return options;
};

export const previousNumberDefault = (previous: unknown, fallback: number | string): number | string => {
	return previousNumber(previous, fallback);
};

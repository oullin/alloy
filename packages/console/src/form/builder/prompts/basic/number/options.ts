import { previousNumberDefault } from '#console/form/builder/prompts/basic/number/defaults';
import { hasPreviousResponse } from '#console/form/builder/validators/previous';
import type { NumberPromptOptions } from '#console/types';

type LabelNumberOptionsInput = {
	defaultValue: number | string;
	hasLabelDefault: boolean;
	hint: string;
	label: string;
	max?: number;
	min?: number;
	placeholder: string;
	previous: unknown;
	required: NumberPromptOptions['required'];
	step?: number;
	transform: NumberPromptOptions['transform'];
	validate: NumberPromptOptions['validate'];
};

export const labelNumberOptions = ({
	defaultValue,
	hasLabelDefault,
	hint,
	label,
	max,
	min,
	placeholder,
	previous,
	required,
	step,
	transform,
	validate,
}: LabelNumberOptionsInput): NumberPromptOptions => {
	const options: NumberPromptOptions = {
		message: label,
		label,
		placeholder,
		required,
		validate,
		hint,
		min,
		max,
		step,
		transform,
	};

	if (hasPreviousResponse(previous)) {
		options.default = previousNumberDefault(previous, hasLabelDefault ? defaultValue : '');
	} else if (hasLabelDefault) {
		options.default = previousNumberDefault(previous, defaultValue);
	}

	return options;
};

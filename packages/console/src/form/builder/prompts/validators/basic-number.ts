import { isBasicPromptLabel, isBasicPromptOptions, parseBasicStepName } from '#console/form/builder/prompts/validators/basic-label';
import type { NumberPromptOptions } from '#console/types';

export type ResolvedNumberFormArguments =
	| {
			kind: 'options';
			name?: string;
			options: NumberPromptOptions;
	  }
	| {
			defaultValue: number | string;
			hasLabelDefault: boolean;
			hint: string;
			kind: 'label';
			label: string;
			max?: number;
			min?: number;
			name?: string;
			placeholder: string;
			required: NumberPromptOptions['required'];
			step?: number;
			transform: NumberPromptOptions['transform'];
			validate: NumberPromptOptions['validate'];
	  };

export const resolveNumberFormArguments = (
	optionsOrLabel: NumberPromptOptions | string,
	placeholder = '',
	defaultValue: number | string = '',
	required: NumberPromptOptions['required'] = false,
	validate: NumberPromptOptions['validate'] = undefined,
	hint = '',
	min?: number,
	max?: number,
	step?: number,
	name?: string,
	transform: NumberPromptOptions['transform'] = undefined,
	argumentCount = 0,
): ResolvedNumberFormArguments => {
	if (isBasicPromptOptions(optionsOrLabel)) {
		return {
			kind: 'options',
			name: parseBasicStepName(placeholder),
			options: optionsOrLabel,
		};
	}

	return {
		defaultValue,
		hasLabelDefault: isBasicPromptLabel(optionsOrLabel) && argumentCount >= 3 && defaultValue !== undefined,
		hint,
		kind: 'label',
		label: optionsOrLabel,
		max,
		min,
		name,
		placeholder,
		required,
		step,
		transform,
		validate,
	};
};

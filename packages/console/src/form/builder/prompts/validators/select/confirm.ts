import { isSelectPromptOptions, parseConfirmDefault, parseSelectStepName } from '#console/form/builder/prompts/validators/select/common';
import type { ConfirmPromptOptions } from '#console/types';

export type ResolvedConfirmFormArguments = {
	name?: string;
	options: ConfirmPromptOptions;
};

export const resolveConfirmFormArguments = (
	optionsOrLabel: ConfirmPromptOptions | string,
	defaultValueOrName: boolean | string = true,
	yes = 'Yes',
	no = 'No',
	required: boolean | string = false,
	validate: ConfirmPromptOptions['validate'] = undefined,
	hint = '',
	name?: string,
	transform: ConfirmPromptOptions['transform'] = undefined,
): ResolvedConfirmFormArguments => {
	if (isSelectPromptOptions(optionsOrLabel)) {
		return {
			name: parseSelectStepName(defaultValueOrName),
			options: optionsOrLabel,
		};
	}

	return {
		name,
		options: {
			message: optionsOrLabel,
			label: optionsOrLabel,
			default: parseConfirmDefault(defaultValueOrName, true),
			yes,
			no,
			required,
			validate,
			hint,
			transform,
		},
	};
};

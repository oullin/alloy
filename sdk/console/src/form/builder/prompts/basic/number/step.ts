import { number } from '#console/prompts/basic';
import { numberOptionsWithPreviousDefault } from '#console/form/builder/prompts/basic/number/defaults';
import type { NumberPromptOptions } from '#console/types';

export const runObjectNumberFormStep = (options: NumberPromptOptions, previous: unknown): Promise<number | string> => {
	return number(numberOptionsWithPreviousDefault(options, previous));
};

export const runLabelNumberFormStep = (options: NumberPromptOptions): Promise<number | string> => {
	return number(options);
};

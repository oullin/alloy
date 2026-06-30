import { promptEnvironment } from '#console/environment';
import { readNumberFallbackValue } from '#console/prompts/number/fallback';
import { readInteractiveNumberValue } from '#console/prompts/number/interactive';
import type { NumberInputOptions, NumberReadResult } from '#console/prompts/number/types';

export const readNumberValue = async (message: string, options: NumberInputOptions = {}): Promise<NumberReadResult> => {
	const environment = promptEnvironment();

	if (!environment.input.readKey) {
		return readNumberFallbackValue(message, options);
	}

	return readInteractiveNumberValue(message, options);
};

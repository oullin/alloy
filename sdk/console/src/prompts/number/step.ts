import { parseNumberStep } from '#console/prompts/number/validators/step';
import { parseNumericValue } from '#console/prompts/number/validators/value';
import type { NumberInputOptions } from '#console/prompts/number/types';

const clamp = (value: number, min?: number, max?: number): number => {
	const clampedMin = min === undefined ? value : Math.max(min, value);

	return max === undefined ? clampedMin : Math.min(max, clampedMin);
};

export const numberStep = (step?: number, integer?: boolean): number => {
	return parseNumberStep(step, integer);
};

export const steppedNumberValue = (value: string, direction: 1 | -1, options: NumberInputOptions): string => {
	const step = numberStep(options.step, options.integer);
	const numeric = parseNumericValue(value);

	if (value === '') {
		const startValue = direction === 1 ? (options.min ?? 1) : (options.max ?? 0);
		const clampedValue = options.min !== undefined ? Math.max(options.min, startValue) : startValue;

		return String(options.max !== undefined ? Math.min(options.max, clampedValue) : clampedValue);
	}

	if (numeric === null) {
		return value;
	}

	const steppedValue = clamp(numeric + step * direction, options.min, options.max);

	return String(options.integer === false ? steppedValue : Math.trunc(steppedValue));
};

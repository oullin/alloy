import { applyTypedKey } from '#console/typed-value';
import { steppedNumberValue } from '#console/prompts/number/step';
import type { NumberInputOptions } from '#console/prompts/number/types';
import type { TypedValueState } from '#console/typed-value/types';

export type NumberReaderState = TypedValueState;

export type NumberReaderInputResult = {
	cancelled: boolean;
	state: NumberReaderState;
	submitted: boolean;
};

export const initialNumberReaderState = (options: NumberInputOptions): NumberReaderState => ({
	cursor: options.hasDefault ? String(options.default).length : 0,
	value: options.hasDefault ? String(options.default) : '',
});

export const applyNumberReaderInput = (state: NumberReaderState, key: string): NumberReaderInputResult => {
	const next = applyTypedKey(state, key);

	if (next.cancelled) {
		return { cancelled: true, state, submitted: false };
	}

	return {
		cancelled: false,
		state: {
			cursor: next.cursor,
			value: next.value,
		},
		submitted: next.submitted,
	};
};

export const steppedNumberReaderState = (state: NumberReaderState, delta: 1 | -1, options: NumberInputOptions): NumberReaderState => {
	const value = steppedNumberValue(state.value, delta, options);

	return {
		cursor: value.length,
		value,
	};
};

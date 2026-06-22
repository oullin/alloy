import { eraseRenderedFrame } from '#console/status/frame';
import { renderCancelledNumberValue, renderNumberValue } from '#console/prompts/number/render';
import { applyNumberReaderInput, initialNumberReaderState, steppedNumberReaderState } from '#console/prompts/number/session/state';
import type { NumberInputOptions } from '#console/prompts/number/types';

export type NumberReaderSession = {
	applyTypedInput(key: string): { cancelled: boolean; submitted: boolean };
	cancel(): void;
	frame(): string;
	step(delta: 1 | -1): void;
	value(): string;
};

export const createNumberReaderSession = (message: string, options: NumberInputOptions): NumberReaderSession => {
	let state = initialNumberReaderState(options);

	let frame = renderNumberValue(message, state.value, state.cursor, options);

	const redraw = (): void => {
		eraseRenderedFrame(frame);
		frame = renderNumberValue(message, state.value, state.cursor, options);
	};

	return {
		applyTypedInput(key) {
			const next = applyNumberReaderInput(state, key);

			if (next.cancelled) {
				return { cancelled: true, submitted: false };
			}

			state = next.state;

			if (!next.submitted) {
				redraw();
			}

			return { cancelled: false, submitted: next.submitted };
		},
		cancel() {
			eraseRenderedFrame(frame);
			renderCancelledNumberValue(message, state.value, options);
		},
		frame() {
			return frame;
		},
		step(delta) {
			state = steppedNumberReaderState(state, delta, options);
			redraw();
		},
		value() {
			return state.value;
		},
	};
};

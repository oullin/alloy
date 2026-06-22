import { eraseRenderedFrame } from '#console/status/frame';
import { applyTypedKey } from '#console/typed-value';
import { passwordLength, renderCancelledPasswordValue, renderPasswordValue } from '#console/prompts/password/render';
import type { PasswordInputOptions } from '#console/prompts/password/types';

export type PasswordReaderSession = {
	applyTypedInput(key: string): { cancelled: boolean; submitted: boolean };
	cancel(): void;
	frame(): string;
	value(): string;
};

export const createPasswordReaderSession = (message: string, options: PasswordInputOptions): PasswordReaderSession => {
	let state = {
		cursor: passwordLength(options.default ?? ''),
		value: options.default ?? '',
	};

	let frame = renderPasswordValue(message, state.value, state.cursor, options);

	const redraw = (): void => {
		eraseRenderedFrame(frame);
		frame = renderPasswordValue(message, state.value, state.cursor, options);
	};

	return {
		applyTypedInput(key) {
			const next = applyTypedKey(state, key);

			if (next.cancelled) {
				return { cancelled: true, submitted: false };
			}

			state = {
				cursor: next.cursor,
				value: next.value,
			};

			if (!next.submitted) {
				redraw();
			}

			return { cancelled: false, submitted: next.submitted };
		},
		cancel() {
			eraseRenderedFrame(frame);
			renderCancelledPasswordValue(message, state.value, options);
		},
		frame() {
			return frame;
		},
		value() {
			return state.value;
		},
	};
};

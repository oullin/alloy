import { promptEnvironment } from '#console/environment';
import { cancelTypedValueRead } from '#console/typed-value/read/cancel';
import { readTypedValueFallback } from '#console/typed-value/read/fallback';
import { eraseRenderedFrame } from '#console/status/frame';
import { applyTypedKey, initialTypedValueState } from '#console/typed-value/edit';
import { renderTypedValue } from '#console/typed-value/render';
import { TEXTAREA_CONTENT_WIDTH } from '#console/typed-value/textarea';
import type { TypedValueOptions, TypedValueState } from '#console/typed-value/types';

export type TypedValueReadResult = {
	cancelled: boolean;
	frame?: string;
	value: string;
};

export const readTypedValue = async (message: string, options: TypedValueOptions = {}): Promise<TypedValueReadResult> => {
	const environment = promptEnvironment();

	if (!environment.input.readKey) {
		return readTypedValueFallback(environment, message, options);
	}

	let state: TypedValueState = initialTypedValueState(options.default ?? '');
	let frame = renderTypedValue(message, state, options);

	while (true) {
		const key = await environment.input.readKey();

		if (key === null) {
			return {
				cancelled: false,
				frame,
				value: state.value,
			};
		}

		const next = applyTypedKey(state, key, options.allowNewLine, options.allowNewLine ? TEXTAREA_CONTENT_WIDTH : undefined);

		if (next.cancelled) {
			return cancelTypedValueRead({ frame, message, options, value: state.value });
		}

		state = {
			cursor: next.cursor,
			value: next.value,
		};

		if (next.submitted) {
			return {
				cancelled: false,
				frame,
				value: state.value,
			};
		}

		eraseRenderedFrame(frame);
		frame = renderTypedValue(message, state, options);
	}
};

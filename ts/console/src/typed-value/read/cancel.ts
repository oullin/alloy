import { cancelPrompt } from '#console/prompt';
import { eraseRenderedFrame } from '#console/status/frame';
import { renderCancelledTypedValue } from '#console/typed-value/render';
import { renderCancelledTextareaFrame } from '#console/typed-value/textarea-frame';
import type { TypedValueReadResult } from '#console/typed-value/read';
import type { TypedValueOptions } from '#console/typed-value/types';

type CancelTypedValueReadOptions = {
	frame: string;
	message: string;
	options: TypedValueOptions;
	value: string;
};

export const cancelTypedValueRead = async ({ frame, message, options, value }: CancelTypedValueReadOptions): Promise<TypedValueReadResult> => {
	eraseRenderedFrame(frame);

	if (!options.allowNewLine) {
		renderCancelledTypedValue(message, value, options);
	} else {
		renderCancelledTextareaFrame(message, value, options);
	}

	return {
		cancelled: true,
		value: await cancelPrompt(value),
	};
};

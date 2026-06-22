import { promptEnvironment } from '#console/environment';
import { renderBox } from '#console/theme/box';
import { cyan, dim, red, strikethrough } from '#console/theme/styles';
import { valueWithCursor } from '#console/typed-value/cursor';
import { visibleTextWindow } from '#console/typed-value/lines';
import { renderTextareaFrame } from '#console/typed-value/textarea-frame';
import type { TypedValueOptions, TypedValueState } from '#console/typed-value/types';

export const renderTypedValue = (message: string, state: TypedValueState, options: TypedValueOptions): string => {
	if (options.allowNewLine) {
		const frame = `${renderTextareaFrame(message, state, options)}\n`;

		promptEnvironment().output.write(frame);

		if (options.hint) {
			promptEnvironment().output.write(`${dim(options.hint)}\n`);
		}

		return options.hint ? `${frame}${dim(options.hint)}\n` : frame;
	}

	const visible = visibleTextWindow(state.value, state.cursor, options.rows);
	const displayValue = state.value.length > 0 ? valueWithCursor(visible.text, visible.cursor) : dim(options.placeholder ?? '');

	const frame = `${renderBox({ body: displayValue, borderStyle: cyan, info: options.hint, title: cyan(message) })}\n`;

	promptEnvironment().output.write(frame);

	return frame;
};

export const renderSubmittedTypedValue = (message: string, value: string): void => {
	promptEnvironment().output.write(`${renderBox({ body: value, title: dim(message) })}\n`);
};

export const renderCancelledTypedValue = (message: string, value: string, options: TypedValueOptions): void => {
	const environment = promptEnvironment();
	const displayValue = value.length > 0 ? value : (options.placeholder ?? '');

	environment.output.write(`${renderBox({ body: strikethrough(dim(displayValue)), borderStyle: red, title: message })}\n`);
	environment.error.write(`${red('  ⚠ Cancelled.')}\n`);
};

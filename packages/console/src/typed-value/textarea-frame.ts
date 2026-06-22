import { renderScrollbarRows } from '#console/concerns/scrollbar';
import { promptEnvironment } from '#console/environment';
import { renderBox } from '#console/theme/box';
import { cyan, dim, red, strikethrough } from '#console/theme/styles';
import { valueWithCursor } from '#console/typed-value/cursor';
import { visibleTextWindow } from '#console/typed-value/lines';
import { TEXTAREA_CONTENT_WIDTH } from '#console/typed-value/textarea';
import { parseTypedValueRows } from '#console/typed-value/validators/rows';
import type { TypedValueOptions, TypedValueState } from '#console/typed-value/types';

export const renderTextareaFrame = (message: string, state: TypedValueState, options: TypedValueOptions): string => {
	return renderBox({ body: textareaBody(state, options), borderStyle: cyan, info: 'Ctrl+D to submit', title: cyan(message) });
};

export const renderSubmittedTextareaFrame = (message: string, value: string): void => {
	promptEnvironment().output.write(`${renderBox({ body: value, title: dim(message) })}\n`);
};

export const renderCancelledTextareaFrame = (message: string, value: string, options: TypedValueOptions): void => {
	const environment = promptEnvironment();
	const body = textareaCancelledBody(value, options);

	environment.output.write(`${renderBox({ body, borderStyle: red, title: message })}\n`);
	environment.error.write(`${red('  ⚠ Cancelled.')}\n`);
};

const textareaBody = (state: TypedValueState, options: TypedValueOptions): string => {
	const rows = parseTypedValueRows(options.rows);

	if (state.value.length === 0) {
		return placeholderBody(options, rows);
	}

	const window = visibleTextWindow(state.value, state.cursor, rows, TEXTAREA_CONTENT_WIDTH);

	if (rows === undefined) {
		return valueWithCursor(window.text, window.cursor);
	}

	const visibleLines = valueWithCursor(window.text, window.cursor).split('\n');
	const padded = [...visibleLines.slice(0, rows), ...Array.from({ length: Math.max(0, rows - visibleLines.length) }, () => '')];

	return renderScrollbarRows(padded, window.start, rows, window.total).join('\n');
};

const placeholderBody = (options: TypedValueOptions, rows: number | undefined): string => {
	const lines = [dim(options.placeholder ?? '')];

	if (rows === undefined) {
		return lines.join('\n');
	}

	return [...lines, ...Array.from({ length: Math.max(0, rows - lines.length) }, () => '')].join('\n');
};

const textareaCancelledBody = (value: string, options: TypedValueOptions): string => {
	const text = value.length > 0 ? value : (options.placeholder ?? '');
	const lines = text.split(/\r?\n/u);

	return lines.map((line) => strikethrough(dim(line))).join('\n');
};

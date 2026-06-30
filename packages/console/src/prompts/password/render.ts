import { promptEnvironment } from '#console/environment';
import { renderBox } from '#console/theme/box';
import { cyan, dim, red, strikethrough } from '#console/theme/styles';
import { characterLength } from '#console/typed-value/characters';
import { valueWithCursor } from '#console/typed-value/cursor';
import type { PasswordInputOptions } from '#console/prompts/password/types';

export const maskPassword = (value: string): string => '•'.repeat(characterLength(value));

export const passwordLength = (value: string): number => characterLength(value);

const passwordDisplay = (value: string, cursor: number, options: PasswordInputOptions): string => {
	return value.length > 0 ? valueWithCursor(maskPassword(value), cursor) : dim(options.placeholder ?? '');
};

export const renderPasswordValue = (message: string, value: string, cursor: number, options: PasswordInputOptions): string => {
	const frame = `${renderBox({ body: passwordDisplay(value, cursor, options), borderStyle: cyan, info: options.hint, title: cyan(message) })}\n`;

	promptEnvironment().output.write(frame);

	return frame;
};

export const renderSubmittedPasswordValue = (message: string, value: string): void => {
	promptEnvironment().output.write(`${renderBox({ body: maskPassword(value), title: dim(message) })}\n`);
};

export const renderCancelledPasswordValue = (message: string, value: string, options: PasswordInputOptions): void => {
	const environment = promptEnvironment();
	const displayValue = value.length > 0 ? maskPassword(value) : (options.placeholder ?? '');

	environment.output.write(`${renderBox({ body: strikethrough(dim(displayValue)), borderStyle: red, title: message })}\n`);
	environment.error.write(`${red('  ⚠ Cancelled.')}\n`);
};

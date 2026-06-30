import { promptEnvironment } from '#console/environment';
import { defaultBackgroundColor, defaultForegroundColor, parseTerminalColor, terminalSupportsTrueColor } from '#console/terminal/capabilities';
import type { TerminalColor } from '#console/terminal/capabilities';
import { parseTerminalDimension, parseTerminalLineCount } from '#console/terminal/validators/size';
import { parseTerminalTitle } from '#console/terminal/validators/title';

export type TerminalSize = {
	columns: number;
	rows: number;
};

export const terminalSize = (): TerminalSize => ({
	columns: parseTerminalDimension(process.stdout.columns, 80),
	rows: parseTerminalDimension(process.stdout.rows, 24),
});

export const clearTerminal = (): void => {
	promptEnvironment().output.write('\u001B[H\u001B[J');
};

export const setTerminalTitle = (title: string): void => {
	promptEnvironment().output.write(`\u001B]0;${parseTerminalTitle(title)}\u0007`);
};

export const eraseLine = (): void => {
	promptEnvironment().output.write('\u001B[2K');
};

export const erasePreviousLines = (count: number): void => {
	const lines = parseTerminalLineCount(count);

	if (lines === 0) {
		return;
	}

	promptEnvironment().output.write('\u001B[1A\u001B[2K'.repeat(lines));
};

export const cursorToStart = (): void => {
	promptEnvironment().output.write('\r');
};

export const hideCursor = (): void => {
	promptEnvironment().output.write('\u001B[?25l');
};

export const showCursor = (): void => {
	promptEnvironment().output.write('\u001B[?25h');
};

export const supportsTrueColor = (value?: string): boolean => {
	return terminalSupportsTrueColor(value);
};

export const foregroundColor = (value?: unknown): TerminalColor => {
	return parseTerminalColor(value, defaultForegroundColor());
};

export const backgroundColor = (value?: unknown): TerminalColor => {
	return parseTerminalColor(value, defaultBackgroundColor());
};

import { parseTerminalColor } from '#console/terminal/validators/color';

export type TerminalColor = readonly [number, number, number];

const TRUE_COLOR_TERMS = new Set(['truecolor', '24bit']);

export const defaultForegroundColor = (): TerminalColor => [204, 204, 204];

export const defaultBackgroundColor = (): TerminalColor => [0, 0, 0];

export const terminalSupportsTrueColor = (value = process.env.COLORTERM): boolean => {
	return TRUE_COLOR_TERMS.has((value ?? '').trim().toLowerCase());
};
export { parseTerminalColor };

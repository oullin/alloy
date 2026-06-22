export const ESC = '\u001B';

const ANSI_SEQUENCE = new RegExp(`^${ESC}\\[[0-9;]*m$`, 'u');
const ANSI_RESET_CODES = new Set(['0', '22', '23', '24', '27', '29', '39', '49']);

export const ansiCode = (sequence: string): string => sequence.slice(2, -1);

export const ansiCodeParts = (code: string): string[] => code.split(';').filter((part) => part.length > 0);

export const isAnsiStyleSequence = (sequence: string): boolean => ANSI_SEQUENCE.test(sequence);

export const isAnsiResetCode = (code: string): boolean => ansiCodeParts(code).some((part) => ANSI_RESET_CODES.has(part));

export const ansiCloseSequence = (codes: string): string => {
	const code = ansiCode(codes);

	if (!isAnsiStyleSequence(codes)) {
		return `${ESC}[0m`;
	}

	if (code === '1' || code === '2') {
		return `${ESC}[22m`;
	}

	if (code === '3') {
		return `${ESC}[23m`;
	}

	if (code === '4') {
		return `${ESC}[24m`;
	}

	if (code === '7') {
		return `${ESC}[27m`;
	}

	if (code === '9') {
		return `${ESC}[29m`;
	}

	return `${ESC}[0m`;
};

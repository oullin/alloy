const ansi = (open: number, close: number, value: string): string => {
	if (value.length === 0) {
		return '';
	}

	return `\u001B[${open}m${value}\u001B[${close}m`;
};

export const dim = (value: string): string => ansi(2, 22, value);

export const cyan = (value: string): string => ansi(36, 39, value);

export const green = (value: string): string => ansi(32, 39, value);

export const inverse = (value: string): string => ansi(7, 27, value);

export const black = (value: string): string => ansi(30, 39, value);

export const red = (value: string): string => ansi(31, 39, value);

export const strikethrough = (value: string): string => ansi(9, 29, value);

export const white = (value: string): string => ansi(37, 39, value);

export const yellow = (value: string): string => ansi(33, 39, value);

export const backgroundCyan = (value: string): string => ansi(46, 49, value);

export const backgroundRed = (value: string): string => ansi(41, 49, value);

export const foregroundRgb = (value: string, red: number, green: number, blue: number): string => {
	if (value.length === 0) {
		return '';
	}

	return `\u001B[38;2;${red};${green};${blue}m${value}\u001B[0m`;
};

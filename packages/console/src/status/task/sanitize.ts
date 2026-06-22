const cursorReset = `${String.fromCharCode(27)}[1G`;
const eraseLine = `${String.fromCharCode(27)}[2K`;

export const sanitizeTaskLine = (line: string): string => {
	return line.replaceAll(cursorReset, '').replaceAll(eraseLine, '').replaceAll('\r', '');
};

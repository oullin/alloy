import stripAnsi from 'strip-ansi';
import { restoreAnsiWrappedLines } from '#console/string-utils/wrap/ansi';
import { plainWrap } from '#console/string-utils/wrap/plain';

export const wrap = (value: string, width: number): string[] => {
	return restoreAnsiWrappedLines(value, plainWrap(stripAnsi(value), width));
};

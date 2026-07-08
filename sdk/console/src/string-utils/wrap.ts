import stripAnsi from 'strip-ansi';
import { restoreAnsiWrappedLines } from '#console/string-utils/wrap/ansi';
import { plainWrap } from '#console/string-utils/wrap/plain';

/** Wraps a string to the given visible width, preserving ANSI styling across lines. */
export const wrap = (value: string, width: number): string[] => {
	return restoreAnsiWrappedLines(value, plainWrap(stripAnsi(value), width));
};

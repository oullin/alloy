import stripAnsi from 'strip-ansi';

export { ansiCloseSequence, ESC } from '#console/string-utils/ansi/codes';
export { parseAnsiSegments } from '#console/string-utils/ansi/segments';
export type { AnsiSegment } from '#console/string-utils/ansi/segments';

export const parseAnsiText = (value: string): string => stripAnsi(value);

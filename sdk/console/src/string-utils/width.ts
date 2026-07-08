import stringWidth from 'string-width';
import stripAnsi from 'strip-ansi';

/** Returns the rendered terminal width of a string, ignoring ANSI escape codes. */
export const visibleWidth = (value: string): number => stringWidth(stripAnsi(value));

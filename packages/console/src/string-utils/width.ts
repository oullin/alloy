import stringWidth from 'string-width';
import stripAnsi from 'strip-ansi';

export const visibleWidth = (value: string): number => stringWidth(stripAnsi(value));

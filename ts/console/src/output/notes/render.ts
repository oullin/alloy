import { visibleWidth } from '#console/strings';
import { parseNoteMessageLines } from '#console/output/notes/validators/message';
import { backgroundCyan, backgroundRed, black, green, red, white, yellow } from '#console/theme/styles';
import type { NoteType } from '#console/output/notes/types';

const paddedIntroLines = (lines: string[]): string[] => {
	const padded = lines.map((line) => ` ${line} `);
	const longest = Math.max(...padded.map(visibleWidth));

	return padded.map((line) => `${line}${' '.repeat(Math.max(0, longest - visibleWidth(line)))}`);
};

const renderNoteLine = (line: string, type?: NoteType | null): string => {
	switch (type) {
		case 'intro':

		case 'outro':
			return ` ${backgroundCyan(black(line))}`;

		case 'warning':
			return yellow(` ${line}`);

		case 'error':
			return red(` ${line}`);

		case 'alert':
			return ` ${backgroundRed(white(` ${line} `))}`;

		case 'info':
			return green(` ${line}`);

		default:
			return ` ${line}`;
	}
};

export const renderNoteLines = (message: string, type?: NoteType | null): string[] => {
	const lines = parseNoteMessageLines(message);

	switch (type) {
		case 'intro':

		case 'outro':
			return paddedIntroLines(lines).map((line) => renderNoteLine(line, type));

		default:
			return lines.map((line) => renderNoteLine(line, type));
	}
};

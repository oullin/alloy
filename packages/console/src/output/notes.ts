import { promptEnvironment } from '#console/environment';
import { renderNoteLines } from '#console/output/notes/render';
import type { NoteType } from '#console/output/notes/types';

export type { NoteType } from '#console/output/notes/types';

export const note = (message: string, type?: NoteType | null): void => {
	promptEnvironment().output.write(`${renderNoteLines(message, type).join('\n')}\n`);
};

export const error = (message: string): void => note(message, 'error');

export const warning = (message: string): void => note(message, 'warning');

export const alert = (message: string): void => note(message, 'alert');

export const info = (message: string): void => note(message, 'info');

export const intro = (message: string): void => note(message, 'intro');

export const outro = (message: string): void => note(message, 'outro');

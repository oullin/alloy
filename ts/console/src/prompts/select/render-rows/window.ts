import { choiceWindow } from '#console/concerns/choices';
import { renderScrollbarRows } from '#console/concerns/scrollbar';
import type { Choice } from '#console/types';

export const renderChoiceWindowRows = <T>(choices: Array<Choice<T>>, selected: number, scroll: number | undefined, renderRow: (choice: Choice<T>, index: number) => string): string => {
	const window = choiceWindow(choices.length, selected, scroll);
	const rows = choices.slice(window.start, window.end).map((choice, offset) => renderRow(choice, window.start + offset));

	return renderScrollbarRows(rows, window.start, window.end - window.start, choices.length).join('\n');
};

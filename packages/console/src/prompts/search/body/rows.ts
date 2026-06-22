import { choiceWindow } from '#console/concerns/choices';
import { choiceLabel } from '#console/concerns/choice-label';
import { renderScrollbarRows } from '#console/concerns/scrollbar';
import { cyan, dim } from '#console/theme/styles';
import type { Choice } from '#console/types';

export const renderSearchRows = <T>(choices: Array<Choice<T>>, highlighted: number | null, marked: Set<number>, scroll: number | undefined, multiple: boolean): string => {
	const window = choiceWindow(choices.length, highlighted ?? 0, scroll);

	const rows = choices.slice(window.start, window.end).map((choice, offset) => {
		const index = window.start + offset;
		const active = highlighted === index;
		const selected = marked.has(index);
		const label = choiceLabel(choice);

		return multiple ? multiSearchRow(label, active, selected) : searchRow(label, active);
	});

	return renderScrollbarRows(rows, window.start, window.end - window.start, choices.length).join('\n');
};

const searchRow = (label: string, active: boolean): string => {
	if (active) {
		return `${cyan('›')} ${label}  `;
	}

	return `  ${dim(label)}  `;
};

const multiSearchRow = (label: string, active: boolean, selected: boolean): string => {
	if (active && selected) {
		return `${cyan('› ◼')} ${label}  `;
	}

	if (active) {
		return `${cyan('›')} ◻ ${label}  `;
	}

	if (selected) {
		return `  ${cyan('◼')} ${dim(label)}  `;
	}

	return `  ${dim('◻')} ${dim(label)}  `;
};

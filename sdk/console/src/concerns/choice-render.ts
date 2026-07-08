import { promptEnvironment } from '#console/environment';
import { choiceWindow } from '#console/concerns/choice-navigation';
import { renderScrollbarRows } from '#console/concerns/scrollbar';
import { choiceDisabledSuffix } from '#console/theme/choice-disabled';
import { cyan, dim } from '#console/theme/styles';
import type { Choice } from '#console/types';

export const renderInteractiveChoices = <T>(message: string, choices: Array<Choice<T>>, selected: number, marked: Set<number> = new Set(), scroll?: number): void => {
	const environment = promptEnvironment();
	const window = choiceWindow(choices.length, selected, scroll);

	environment.output.write(`${message}\n`);

	const rows = choices.slice(window.start, window.end).map((choice, offset) => {
		const index = window.start + offset;
		const pointer = index === selected ? '›' : ' ';
		const checked = marked.size > 0 ? (marked.has(index) ? '[x]' : '[ ]') : '  ';
		const disabled = choiceDisabledSuffix(choice.disabled);
		const hint = choice.hint ? ` ${choice.hint}` : '';

		return `${pointer} ${checked} ${choice.label}${hint}${disabled}`;
	});

	for (const row of renderScrollbarRows(rows, window.start, window.end - window.start, choices.length)) {
		environment.output.write(`${row}\n`);
	}
};

export const renderInteractiveChecklist = <T>(message: string, choices: Array<Choice<T>>, selected: number, marked: Set<number>, scroll?: number): void => {
	const environment = promptEnvironment();
	const window = choiceWindow(choices.length, selected, scroll);

	environment.output.write(`${message}\n`);

	const rows = choices.slice(window.start, window.end).map((choice, offset) => {
		const index = window.start + offset;
		const active = index === selected;
		const checked = marked.has(index);
		const pointer = active ? '›' : ' ';
		const marker = checked ? '◼' : '◻';
		const disabled = choiceDisabledSuffix(choice.disabled);
		const hint = choice.hint ? ` ${choice.hint}` : '';
		const label = `${choice.label}${hint}${disabled}`;

		if (active && checked) {
			return `${cyan(`${pointer} ${marker}`)} ${label}`;
		}

		if (active) {
			return `${cyan(pointer)} ${marker} ${label}`;
		}

		if (checked) {
			return `  ${cyan(marker)} ${dim(label)}`;
		}

		return `  ${dim(marker)} ${dim(label)}`;
	});

	for (const row of renderScrollbarRows(rows, window.start, window.end - window.start, choices.length)) {
		environment.output.write(`${row}\n`);
	}
};

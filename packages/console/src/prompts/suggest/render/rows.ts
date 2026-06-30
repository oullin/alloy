import { choiceWindow } from '#console/concerns/choices';
import { renderScrollbarRows } from '#console/concerns/scrollbar';
import { cyan, dim } from '#console/theme/styles';

export const renderSuggestRows = (matches: string[], highlighted: number | null, scroll?: number): string => {
	const window = choiceWindow(matches.length, highlighted ?? 0, scroll);

	const rows = matches.slice(window.start, window.end).map((match, offset) => {
		const index = window.start + offset;

		if (index === highlighted) {
			return `${cyan('›')} ${match}  `;
		}

		return `  ${dim(match)}  `;
	});

	return renderScrollbarRows(rows, window.start, window.end - window.start, matches.length).join('\n');
};

import { parseScrollbarState } from '#console/concerns/validators/scrollbar';
import { visibleWidth } from '#console/strings';
import { cyan, dim } from '#console/theme/styles';

export const scrollbarPosition = (firstVisible: number, height: number, total: number): number => {
	if (firstVisible === 0) {
		return 0;
	}

	const maxPosition = total - height;

	if (firstVisible === maxPosition) {
		return height - 1;
	}

	if (height <= 2) {
		return -1;
	}

	return Math.round((firstVisible / maxPosition) * (height - 3)) + 1;
};

export const renderScrollbarRows = (rows: string[], firstVisible: number, height: number, total: number, style = cyan): string[] => {
	const state = parseScrollbarState({ firstVisible, height, total });

	if (state === undefined || rows.length === 0) {
		return rows;
	}

	const position = scrollbarPosition(state.firstVisible, state.height, state.total);
	const width = Math.max(...rows.map((row) => visibleWidth(row)));

	return rows.map((row, index) => {
		const bar = index === position ? style('┃') : dim('│');

		return `${padVisible(row, width)} ${bar}`;
	});
};

const padVisible = (value: string, width: number): string => {
	return `${value}${' '.repeat(Math.max(0, width - visibleWidth(value)))}`;
};

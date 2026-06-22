import { visibleWidth } from '#console/strings';
import type { GridLayout } from '#console/output/grid/layout';

const padGridCell = (value: string, width: number): string => `${value}${' '.repeat(Math.max(0, width - visibleWidth(value)))}`;

const gridHorizontalLine = (left: string, middle: string, right: string, widths: number[]): string => {
	return `${left}${widths.map((width) => '─'.repeat(width + 2)).join(middle)}${right}`;
};

export const boxedGridLines = ({ rows, widths }: GridLayout): string[] => {
	const renderedRows: string[] = [gridHorizontalLine('┌', '┬', '┐', widths)];

	rows.forEach((row, index) => {
		if (index > 0) {
			renderedRows.push(gridHorizontalLine('├', '┼', '┤', widths));
		}

		renderedRows.push(`│ ${widths.map((columnWidth, column) => padGridCell(row[column] ?? '', columnWidth)).join(' │ ')} │`);
	});

	renderedRows.push(gridHorizontalLine('└', '┴', '┘', widths));

	return renderedRows;
};

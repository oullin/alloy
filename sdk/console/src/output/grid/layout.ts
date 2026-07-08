import { visibleWidth } from '#console/strings';
import { parseGridWidth } from '#console/output/grid/validators/layout';

export type GridLayout = {
	rows: string[][];
	widths: number[];
};

const balancedColumnCount = (itemCount: number, maxColumns: number): number => {
	if (itemCount <= maxColumns) {
		return itemCount;
	}

	for (let columns = maxColumns; columns >= 1; columns -= 1) {
		const remainder = itemCount % columns;

		if (remainder === 0 || remainder >= Math.ceil(columns / 2)) {
			return columns;
		}
	}

	return maxColumns;
};

const chunkGridItems = (items: string[], columnCount: number): string[][] => {
	const rows: string[][] = [];

	for (let index = 0; index < items.length; index += columnCount) {
		const row = items.slice(index, index + columnCount);

		while (row.length < columnCount) {
			row.push('');
		}

		rows.push(row);
	}

	return rows;
};

export const createGridLayout = (values: string[], availableWidth: number): GridLayout => {
	if (values.length === 0) {
		return {
			rows: [],
			widths: [],
		};
	}

	const cellWidth = Math.max(...values.map(visibleWidth)) + 4;
	const usableWidth = parseGridWidth(availableWidth, 1);
	const maxColumns = Math.max(1, Math.floor((usableWidth - 1) / (cellWidth + 1)));
	const columnCount = Math.max(1, balancedColumnCount(values.length, maxColumns));
	const rows = chunkGridItems(values, columnCount);

	return {
		rows,
		widths: Array.from({ length: columnCount }, (_, column) => {
			return Math.max(...rows.map((row) => visibleWidth(row[column] ?? '')));
		}),
	};
};

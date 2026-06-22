import { parseDataTableColumnCount, parseDataTableWidth } from '#console/output/data-table/widths/validators/dimensions';

export const availableDataTableCellWidth = (columnCount: number, maxWidth: number): number => {
	const columns = parseDataTableColumnCount(columnCount);
	const width = parseDataTableWidth(maxWidth, 1);
	const markerWidth = 1;
	const renderedColumnCount = columns + 1;
	const tableOverhead = renderedColumnCount * 3 + 1;
	const scrollbarArea = 2;
	const frameInset = 6;

	return width - markerWidth - tableOverhead - scrollbarArea - frameInset;
};

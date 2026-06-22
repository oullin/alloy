import { visibleWidth } from '#console/strings';
import { splitMultilineDataTableCell } from '#console/output/data-table/multiline';

export const naturalDataTableColumnWidth = (rows: string[][], column: number): number => {
	const widths = rows
		.map((row) => row[column] ?? '')
		.flatMap((cell) => splitMultilineDataTableCell(cell).map(visibleWidth))
		.filter((width) => width > 0)
		.sort((left, right) => left - right);

	if (widths.length === 0) {
		return 0;
	}

	const percentile = widths[Math.max(0, Math.ceil(widths.length * 0.9) - 1)] ?? 0;
	const maximum = widths.at(-1) ?? 0;

	return maximum <= percentile * 2 ? maximum : percentile;
};

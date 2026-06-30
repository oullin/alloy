import { visibleWidth } from '#console/strings';
import { availableDataTableCellWidth } from '#console/output/data-table/widths/available';
import { naturalDataTableColumnWidth } from '#console/output/data-table/widths/natural';
import { shrinkDataTableColumnWidths } from '#console/output/data-table/widths/shrink';

export const dataTableColumnWidths = (headers: string[], rows: string[][], maxWidth: number): number[] => {
	const columnCount = Math.max(headers.length, ...rows.map((row) => row.length));
	const headerWidths = Array.from({ length: columnCount }, (_, index) => visibleWidth(headers[index] ?? ''));
	const natural = headerWidths.map((headerWidth, index) => Math.max(headerWidth, naturalDataTableColumnWidth(rows, index)));
	const available = availableDataTableCellWidth(columnCount, maxWidth);
	const totalNatural = natural.reduce((sum, width) => sum + width, 0);

	if (available <= 0) {
		return Array.from({ length: columnCount }, () => 1);
	}

	if (totalNatural <= available) {
		return natural;
	}

	return shrinkDataTableColumnWidths({ available, headerWidths, natural });
};

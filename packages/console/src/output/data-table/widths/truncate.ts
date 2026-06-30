import { truncate, visibleWidth } from '#console/strings';
import { splitMultilineDataTableCell } from '#console/output/data-table/multiline';

type TruncateDataTableColumnsOptions = {
	headers: string[];
	rows: string[][];
	widths: number[];
};

export const truncateDataTableColumns = ({ headers, rows, widths }: TruncateDataTableColumnsOptions): { headers: string[]; rows: string[][] } => ({
	headers: headers.map((header, index) => truncate(header, widths[index] ?? visibleWidth(header))),
	rows: rows.map((row) => row.map((cell, index) => truncateDataTableCell(cell, widths[index] ?? visibleWidth(cell)))),
});

const truncateDataTableCell = (cell: string, width: number): string =>
	splitMultilineDataTableCell(cell)
		.map((line) => truncate(line, width))
		.join('\n');

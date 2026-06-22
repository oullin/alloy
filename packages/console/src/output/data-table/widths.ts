import { terminalSize } from '#console/terminal';
import { truncateDataTableColumns } from '#console/output/data-table/widths/truncate';
import { dataTableColumnWidths } from '#console/output/data-table/widths/columns';
import { parseDataTableWidth } from '#console/output/data-table/widths/validators/dimensions';

type FitDataTableColumnsOptions = {
	allRows: string[][];
	headers: string[];
	maxWidth?: number;
	rows: string[][];
};

export const fitDataTableColumns = ({ allRows, headers, maxWidth = terminalSize().columns, rows }: FitDataTableColumnsOptions): { headers: string[]; rows: string[][] } => {
	const widths = dataTableColumnWidths(headers, allRows, parseDataTableWidth(maxWidth, terminalSize().columns));

	return truncateDataTableColumns({ headers, rows, widths });
};

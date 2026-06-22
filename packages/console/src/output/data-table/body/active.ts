import { renderTable } from '#console/theme';
import { expandMultilineDataTableRows } from '#console/output/data-table/multiline';
import { dataTableRowCells } from '#console/output/data-table/rows';
import { renderDataTableHeaders } from '#console/output/data-table/body/headers';
import { renderScrollableDataTable } from '#console/output/data-table/scrollbar';
import { dataTableRowWindow } from '#console/output/data-table/selection';
import { fixedVisualDataTableRows } from '#console/output/data-table/visual-window';
import { fitDataTableColumns } from '#console/output/data-table/widths';
import type { DataTableBodyOptions } from '#console/output/data-table/body/types';

export const renderDataTableBody = <T>(options: DataTableBodyOptions<T>): string[] => {
	if (options.rows.length === 0) {
		return [renderTable([], [['No results found.']])];
	}

	const window = dataTableRowWindow(options.rows.length, options.selected, options.scroll);
	const visibleCells = options.rows.slice(window.start, window.end).map(({ row }) => dataTableRowCells(options.headers, row));
	const allCells = options.allRows.map((row) => dataTableRowCells(options.headers, row));
	const fitted = fitDataTableColumns({ allRows: allCells, headers: options.headers, rows: visibleCells });
	const visualRows = fixedVisualDataTableRows(expandMultilineDataTableRows(renderSelectedDataTableRows(fitted.rows, window.start, options.selected)), options.scroll);
	const lines = [renderScrollableDataTable(renderDataTableHeaders(fitted.headers), visualRows, window.start, window.end - window.start, options.rows.length)];

	if (window.end - window.start < options.rows.length) {
		const suffix = options.query.length > 0 ? ' results' : '';

		lines.push(`  Viewing ${window.start + 1}-${window.end} of ${options.rows.length}${suffix}`);
	}

	return lines;
};

const renderSelectedDataTableRows = (rows: string[][], start: number, selected: number): string[][] => {
	return rows.map((row, offset) => {
		const index = start + offset;

		return [index === selected ? '›' : ' ', ...row];
	});
};

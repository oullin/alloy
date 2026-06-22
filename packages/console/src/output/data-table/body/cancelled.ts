import { expandMultilineDataTableRows } from '#console/output/data-table/multiline';
import { dataTableRowCells } from '#console/output/data-table/rows';
import { dim, strikethrough } from '#console/theme/styles';
import type { VisibleDataTableRow } from '#console/output/data-table/types';

export const renderCancelledDataTableRows = <T>(headers: string[], rows: Array<VisibleDataTableRow<T>>, selected: number): string[][] => {
	return rows.flatMap(({ row }, index) => {
		const expandedRows = expandMultilineDataTableRows([[index === selected ? '›' : ' ', ...dataTableRowCells(headers, row)]]);

		return expandedRows.map((expandedRow) => {
			return expandedRow.map((cell, columnIndex) => {
				if (columnIndex === 0) {
					return cell;
				}

				return cell.length > 0 ? dim(strikethrough(cell)) : '';
			});
		});
	});
};

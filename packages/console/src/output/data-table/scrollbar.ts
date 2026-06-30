import { renderScrollbarRows } from '#console/concerns/scrollbar';
import { renderTable } from '#console/theme';

export const renderScrollableDataTable = (headers: string[], rows: string[][], firstVisible: number, visibleRows: number, totalRows: number): string => {
	const lines = renderTable(headers, rows).split('\n');
	const headerLineCount = headers.length > 0 ? 2 : 0;
	const head = lines.slice(0, headerLineCount);
	const body = lines.slice(headerLineCount);

	return [...head, ...renderScrollbarRows(body, firstVisible, visibleRows, totalRows)].join('\n');
};

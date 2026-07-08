import { visibleWidth } from '#console/strings';
import { isRenderableTableColumnCount } from '#console/theme/validators/table';

const padVisible = (value: string, width: number): string => `${value}${' '.repeat(Math.max(0, width - visibleWidth(value)))}`;

const tableColumnCount = (headers: string[], rows: string[][]): number => {
	return Math.max(headers.length, ...rows.map((row) => row.length));
};

export const renderTable = (headers: string[], rows: string[][]): string => {
	const columnCount = tableColumnCount(headers, rows);

	if (!isRenderableTableColumnCount(columnCount)) {
		return '';
	}

	const widths = Array.from({ length: columnCount }, (_, index) => {
		return Math.max(visibleWidth(headers[index] ?? ''), ...rows.map((row) => visibleWidth(row[index] ?? '')));
	});

	const renderRow = (columns: string[]): string => {
		return `| ${widths.map((width, index) => padVisible(columns[index] ?? '', width)).join(' | ')} |`;
	};

	if (headers.length === 0) {
		return rows.map(renderRow).join('\n');
	}

	return [renderRow(headers), renderRow(widths.map((width) => '-'.repeat(width))), ...rows.map(renderRow)].join('\n');
};

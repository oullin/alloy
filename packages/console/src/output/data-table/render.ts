import { promptEnvironment } from '#console/environment';
import { renderTable } from '#console/theme';
import { renderCancelledDataTableRows, renderDataTableBody, renderDataTableHeaders, renderSubmittedDataTableRow } from '#console/output/data-table/render-body';
import { clampDataTableSelection } from '#console/output/data-table/selection';
import { dim, red } from '#console/theme/styles';
import type { DataTableRow } from '#console/types';
import type { DataTableSearchMode } from '#console/output/data-table/search';
import type { VisibleDataTableRow } from '#console/output/data-table/types';

type RenderDataTableFrameOptions<T> = {
	allRows: Array<DataTableRow<T>>;
	headers: string[];
	message: string;
	mode: 'browse' | 'search';
	query: string;
	rows: Array<VisibleDataTableRow<T>>;
	scroll?: number;
	selected: number;
};

export type RenderedDataTableFrame = {
	frame: string;
	selected: number;
};

export const renderDataTableFrame = <T>(options: RenderDataTableFrameOptions<T>): RenderedDataTableFrame => {
	const environment = promptEnvironment();
	const selected = clampDataTableSelection(options.selected, options.rows);
	const querySuffix = options.mode === 'search' || options.query.length > 0 ? ` ${options.query}` : '';
	const lines = [`${options.message}${querySuffix}`];

	lines.push(...renderDataTableBody({ ...options, selected }));

	const frame = `${lines.join('\n')}\n`;

	environment.output.write(frame);

	return { frame, selected };
};

export const renderSubmittedDataTableFrame = <T>(message: string, headers: string[], rows: Array<VisibleDataTableRow<T>>, selected: number): void => {
	const display = renderSubmittedDataTableRow(headers, rows, selected);

	if (display === null) {
		return;
	}

	promptEnvironment().output.write(`${message}\n${display}\n`);
};

export const renderCancelledDataTableFrame = <T>(
	message: string,
	headers: string[],
	rows: Array<VisibleDataTableRow<T>>,
	selected: number,
	mode: DataTableSearchMode,
	query: string,
): void => {
	const outputRows = renderCancelledDataTableRows(headers, rows, selected);
	const querySuffix = mode === 'search' || query.length > 0 ? ` ${query}` : '';

	const environment = promptEnvironment();

	environment.output.write(`${message}\n`);
	environment.output.write(`${dim(`/ Search${querySuffix}`)}\n`);
	environment.output.write(`${renderTable(renderDataTableHeaders(headers), outputRows)}\n`);
	environment.error.write(`${red('  ⚠ Cancelled.')}\n`);
};

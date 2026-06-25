import { PromptValidationError } from '#console/prompt';
import { choiceValueEquals } from '#console/concerns/choices';
import { eraseRenderedFrame } from '#console/status/frame';
import { dataTableRowValue } from '#console/output/data-table/rows';
import type { DataTableSelectionReadResult, VisibleDataTableRow } from '#console/output/data-table/types';

const invalidRow = (): PromptValidationError => new PromptValidationError('Please select a valid row.');

export const selectedDataTableValue = <T>(rows: Array<VisibleDataTableRow<T>>, selected: number): T | number => {
	const selectedRow = rows[selected];

	if (!selectedRow) {
		throw invalidRow();
	}

	return dataTableRowValue(selectedRow.row, selectedRow.index);
};

export const initialDataTableSelection = <T>(rows: Array<VisibleDataTableRow<T>>, defaultValue: T | number | undefined, hasDefault = false): number => {
	if (!hasDefault) {
		return 0;
	}

	const selected = rows.findIndex(({ index, row }) => choiceValueEquals(dataTableRowValue(row, index), defaultValue));

	return Math.max(0, selected);
};

export const dataTableSelectionResult = <T>(
	rows: Array<VisibleDataTableRow<T>>,
	selected: number,
	submitted: boolean,
	cancelled = false,
	frame?: string,
): DataTableSelectionReadResult<T> => ({
	cancelled,
	frame,
	rows,
	selected,
	submitted,
	value: selectedDataTableValue(rows, selected),
});

export const assertSelectedDataTableRow = <T>(rows: Array<VisibleDataTableRow<T>>, selected: number, frame: string): void => {
	if (rows[selected]) {
		return;
	}

	if (frame.length > 0) {
		eraseRenderedFrame(frame);
	}

	throw invalidRow();
};

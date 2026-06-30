import type { BasePromptOptions } from '#console/contracts/options';
import type { MaybePromise } from '#console/contracts/base';

export type TableCell = string | number | boolean | null | undefined;

export type TableOptions = {
	headers?: string[];
	rows: Array<Array<TableCell> | Record<string, TableCell>>;
};

export type DataTablePromptOptions<T = unknown> = BasePromptOptions<T | number> & {
	filter?: (query: string, row: DataTableRow<T>) => boolean;
	headers?: string[];
	rows: Array<DataTableRow<T>>;
	scroll?: number;
	transform?: (value: T | number) => MaybePromise<T | number>;
};

export type DataTableObjectRow<T = unknown> = {
	cells: Record<string, TableCell>;
	value?: T;
};

export type DataTableRow<T = unknown> = Array<TableCell> | Record<string, TableCell> | DataTableObjectRow<T>;

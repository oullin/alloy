import type { DataTablePromptOptions } from '#console/types';

export type DataTableReadOptions<T> = DataTablePromptOptions<T> & {
	hasDefault?: boolean;
};

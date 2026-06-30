import type { SearchPromptOptions } from '#console/types';

export type SearchReadOptions<T> = SearchPromptOptions<T> & {
	hasDefault?: boolean;
};

export type SearchReaderSelection<T> = {
	label: string;
	submitted: boolean;
	value: T | undefined;
};

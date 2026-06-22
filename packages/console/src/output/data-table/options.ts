import { parseDataTableDefault, parseInitialDataTableRow } from '#console/output/validators/data-table';
import { dataTableRowValue } from '#console/output/data-table/rows';
import { hasPromptDefault } from '#console/validators/default';
import type { DataTablePromptOptions, DataTableRow } from '#console/types';

export type NormalizedDataTablePromptOptions<T> = DataTablePromptOptions<T> & {
	hasDefault: boolean;
};

export const normalizeDataTablePromptOptions = <T>(options: DataTablePromptOptions<T>): NormalizedDataTablePromptOptions<T> => ({
	...options,
	hasDefault: hasPromptDefault(options),
});

export const transformDataTableValue = async <T>(options: Pick<DataTablePromptOptions<T>, 'transform'>, value: T | number): Promise<T | number> => {
	return options.transform ? options.transform(value) : value;
};

export const preserveDataTableRetryDefault = <T>(options: NormalizedDataTablePromptOptions<T>, value: T | number): void => {
	options.default = value;
	options.hasDefault = true;
};

export const initialDataTableDefault = <T>(rows: Array<DataTableRow<T>>): T | number => {
	return dataTableRowValue(parseInitialDataTableRow<T>(rows), 0);
};

export const dataTableValidationOptions = async <T>(options: NormalizedDataTablePromptOptions<T>, useInitialDefault = false): Promise<DataTablePromptOptions<T>> => {
	const defaultValue = options.hasDefault ? parseDataTableDefault<T>(options.default) : useInitialDefault ? initialDataTableDefault(options.rows) : undefined;

	let transformedDefault = defaultValue;

	if (defaultValue !== undefined) {
		try {
			transformedDefault = await transformDataTableValue(options, defaultValue);
		} catch {
			transformedDefault = defaultValue;
		}
	}

	return {
		...options,
		default: transformedDefault,
	};
};

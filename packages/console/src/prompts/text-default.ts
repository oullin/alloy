import type { TextPromptOptions } from '#console/types';

export const transformedTextDefault = async (options: Pick<TextPromptOptions, 'default' | 'transform'>): Promise<string> => {
	const value = options.default ?? '';

	try {
		return options.transform ? await options.transform(value) : value;
	} catch {
		return value;
	}
};

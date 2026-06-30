import { resolveSuggestionSource } from '#console/prompts/suggest/validators/source';
import type { SuggestOptions } from '#console/prompts/suggest/options';

export const resolveSuggestions = async (source: SuggestOptions['options'], query: string): Promise<string[]> => {
	const resolved = await resolveSuggestionSource(source, query);

	if (!resolved.filter) {
		return [...resolved.options];
	}

	const normalizedQuery = query.toLowerCase();

	return resolved.options.filter((option) => option.toLowerCase().startsWith(normalizedQuery));
};

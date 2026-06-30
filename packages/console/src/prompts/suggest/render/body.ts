import { dim } from '#console/theme/styles';
import { valueWithCursor } from '#console/typed-value/cursor';
import { renderSuggestRows } from '#console/prompts/suggest/render/rows';

export const renderSuggestBody = (value: string, cursor: number, placeholder: string, matches: string[], highlighted: number | null, scroll?: number): string => {
	const query = value.length > 0 ? valueWithCursor(value, cursor) : dim(placeholder);
	const rows = renderSuggestRows(matches, highlighted, scroll);

	if (matches.length === 0) {
		return [query, dim('  No results.')].join('\n');
	}

	return rows.length > 0 ? [query, rows].join('\n') : query;
};

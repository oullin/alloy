import { promptEnvironment } from '#console/environment';
import { resolveInfo } from '#console/concerns/info';
import { renderSuggestBody } from '#console/prompts/suggest/render/body';
import { renderBox } from '#console/theme/box';
import { cyan, dim, red, strikethrough } from '#console/theme/styles';
import type { SuggestOptions } from '#console/prompts/suggest/options';

export const renderSuggestions = (
	message: string,
	value: string,
	cursor: number,
	matches: string[],
	highlighted: number | null,
	scroll?: number,
	info?: SuggestOptions['info'],
	placeholder = '',
): string => {
	const text = resolveInfo(info, highlighted === null ? null : (matches[highlighted] ?? null));
	const frame = `${renderBox({ body: renderSuggestBody(value, cursor, placeholder, matches, highlighted, scroll), borderStyle: cyan, info: text, title: cyan(message) })}\n`;

	promptEnvironment().output.write(frame);

	return frame;
};

export const renderSubmittedSuggestion = (message: string, value: string): void => {
	promptEnvironment().output.write(`${renderBox({ body: value, title: dim(message) })}\n`);
};

export const renderCancelledSuggestion = (message: string, value: string, placeholder = ''): void => {
	const displayValue = value.length > 0 ? value : placeholder;

	promptEnvironment().output.write(`${renderBox({ body: strikethrough(dim(displayValue)), borderStyle: red, title: message })}\n`);
	promptEnvironment().error.write(`${red('  ⚠ Cancelled.')}\n`);
};

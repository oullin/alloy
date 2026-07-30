import { promptEnvironment } from '#console/environment';
import { joinedInfoDetails, resolveInfo } from '#console/concerns/info';
import { renderSearchBody } from '#console/prompts/search/render-body';
import { renderBox } from '#console/theme/box';
import { cyan, dim, red, strikethrough } from '#console/theme/styles';
import type { Choice, SearchPromptOptions } from '#console/types';

export const renderSearchChoices = <T>(
	message: string,
	query: string,
	cursor: number,
	choices: Array<Choice<T>>,
	highlighted: number | null,
	marked: Set<number> = new Set(),
	selectedLabels: string[] = [],
	scroll?: number,
	info?: SearchPromptOptions<T>['info'],
	showSelectedSummary = false,
	placeholder = '',
): string => {
	const text = resolveInfo(info, highlighted === null ? null : (choices[highlighted]?.value ?? null));
	const summary = showSelectedSummary ? selectedSummary(selectedLabels.length, selectedLabels.length - marked.size) : '';
	const details = joinedInfoDetails(text, summary);

	const frame = `${renderBox({
		body: renderSearchBody(query, cursor, placeholder, choices, highlighted, marked, scroll, showSelectedSummary),
		borderStyle: cyan,
		info: details,
		title: cyan(message),
	})}\n`;

	promptEnvironment().output.write(frame);

	return frame;
};

export const renderSubmittedSearchChoice = (message: string, label: string): void => {
	promptEnvironment().output.write(`${renderBox({ body: label, title: dim(message) })}\n`);
};

export const renderSubmittedSearchChoices = (message: string, labels: string[]): void => {
	promptEnvironment().output.write(`${renderBox({ body: labels.join('\n'), title: dim(message) })}\n`);
};

export const renderCancelledSearch = (message: string, query: string, placeholder = ''): void => {
	const body = strikethrough(dim(query.length > 0 ? query : placeholder));

	promptEnvironment().output.write(`${renderBox({ body, borderStyle: red, title: dim(message) })}\n`);
	promptEnvironment().error.write(`${red('  ⚠ Cancelled.')}\n`);
};

const selectedSummary = (selectedCount: number, hiddenCount: number): string => {
	const hidden = hiddenCount > 0 ? ` (${hiddenCount} hidden)` : '';

	return `${selectedCount} selected${hidden}`;
};

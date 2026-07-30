import { promptEnvironment } from '#console/environment';
import { joinedInfoDetails, resolveInfo } from '#console/concerns/info';
import { renderActiveChecklistRows, renderActiveChoiceRows, renderCancelledChecklistRows, renderCancelledChoiceRows } from '#console/prompts/select/render-rows';
import { renderBox } from '#console/theme/box';
import { cyan, dim, red } from '#console/theme/styles';
import type { Choice, MultiSelectPromptOptions, SelectPromptOptions } from '#console/types';

export const renderSelectedChoice = <T>(message: string, choices: Array<Choice<T>>, selected: number, scroll: number | undefined, info: SelectPromptOptions<T>['info']): string => {
	const text = resolveInfo(info, choices[selected]?.value ?? null);

	const frame = `${renderBox({ body: renderActiveChoiceRows(choices, selected, scroll), borderStyle: cyan, info: text, title: cyan(message) })}\n`;

	promptEnvironment().output.write(frame);

	return frame;
};

export const renderSubmittedChoice = (message: string, label: string): void => {
	promptEnvironment().output.write(`${renderBox({ body: label, title: dim(message) })}\n`);
};

export const renderSubmittedChoices = (message: string, labels: string[]): void => {
	promptEnvironment().output.write(`${renderBox({ body: labels.join('\n'), title: dim(message) })}\n`);
};

export const renderCancelledChoice = <T>(message: string, choices: Array<Choice<T>>, selected: number, scroll?: number): void => {
	promptEnvironment().output.write(`${renderBox({ body: renderCancelledChoiceRows(choices, selected, scroll), borderStyle: red, title: message })}\n`);
	promptEnvironment().error.write(`${red('  ⚠ Cancelled.')}\n`);
};

export const renderCancelledChoices = <T>(message: string, choices: Array<Choice<T>>, selected: number, marked: Set<number>, scroll?: number): void => {
	promptEnvironment().output.write(`${renderBox({ body: renderCancelledChecklistRows(choices, selected, marked, scroll), borderStyle: red, title: message })}\n`);
	promptEnvironment().error.write(`${red('  ⚠ Cancelled.')}\n`);
};

export const renderMultipleChoices = <T>(
	message: string,
	choices: Array<Choice<T>>,
	selected: number,
	marked: Set<number>,
	scroll?: number,
	info?: MultiSelectPromptOptions<T>['info'],
): string => {
	const text = resolveInfo(info, choices[selected]?.value ?? null);
	const summary = scroll !== undefined && choices.length > scroll ? `${marked.size} selected` : '';
	const details = joinedInfoDetails(text, summary);

	const frame = `${renderBox({ body: renderActiveChecklistRows(choices, selected, marked, scroll), borderStyle: cyan, info: details, title: cyan(message) })}\n`;

	promptEnvironment().output.write(frame);

	return frame;
};

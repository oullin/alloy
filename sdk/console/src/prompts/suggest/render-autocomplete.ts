import { promptEnvironment } from '#console/environment';
import { joinedInfoDetails, resolveInfo } from '#console/concerns/info';
import { autocompleteDisplayValue } from '#console/prompts/suggest/ghost-text';
import { renderBox } from '#console/theme/box';
import { cyan, dim, red, strikethrough } from '#console/theme/styles';
import type { SuggestOptions } from '#console/prompts/suggest/options';
import type { TypedValueState } from '#console/typed-value/types';

export const renderAutocomplete = (message: string, state: TypedValueState, matches: string[], highlighted: number, hint = '', placeholder = '', info?: SuggestOptions['info']): string => {
	const value = state.value.length === 0 ? dim(placeholder) : autocompleteDisplayValue(state, matches[highlighted], placeholder);

	const details = joinedInfoDetails(hint, resolveInfo(info, matches[highlighted] ?? null));

	const frame = `${renderBox({ body: value, borderStyle: cyan, info: details, title: cyan(message) })}\n`;

	promptEnvironment().output.write(frame);

	return frame;
};

export const renderSubmittedAutocomplete = (message: string, value: string): void => {
	promptEnvironment().output.write(`${renderBox({ body: value, title: dim(message) })}\n`);
};

export const renderCancelledAutocomplete = (message: string, value: string, placeholder = ''): void => {
	const environment = promptEnvironment();
	const displayValue = value.length > 0 ? value : placeholder;

	environment.output.write(`${renderBox({ body: strikethrough(dim(displayValue)), borderStyle: red, title: message })}\n`);
	environment.error.write(`${red('  ⚠ Cancelled.')}\n`);
};

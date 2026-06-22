import { choiceDisabledSuffix } from '#console/theme/choice-disabled';
import type { Choice } from '#console/types';

export const choiceLabel = <T>(choice: Choice<T>): string => {
	const disabled = choiceDisabledSuffix(choice.disabled);
	const hint = choice.hint ? ` ${choice.hint}` : '';

	return `${choice.label}${hint}${disabled}`;
};

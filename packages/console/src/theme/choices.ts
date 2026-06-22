import { choiceDisabledSuffix } from '#console/theme/choice-disabled';
import type { Choice } from '#console/types';

export const renderChoices = <T>(choices: Array<Choice<T>>): string => {
	return choices
		.map((choice, index) => {
			const disabled = choiceDisabledSuffix(choice.disabled);
			const hint = choice.hint ? ` ${choice.hint}` : '';

			return `  ${index + 1}. ${choice.label}${hint}${disabled}`;
		})
		.join('\n');
};

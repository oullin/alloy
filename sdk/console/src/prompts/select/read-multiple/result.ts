import { markedChoiceValues } from '#console/prompts/select/multiple';
import type { MultipleChoicesReadResult } from '#console/prompts/select/read-multiple/types';
import type { Choice } from '#console/types';

export const markedChoiceLabels = <T>(choices: Array<Choice<T>>, marked: Set<number>): string[] => {
	return [...marked]
		.sort((left, right) => left - right)
		.map((index) => choices[index]?.label)
		.filter((label): label is string => label !== undefined);
};

export const multipleChoicesResult = <T>(choices: Array<Choice<T>>, marked: Set<number>, submitted: boolean, cancelled = false, frame?: string): MultipleChoicesReadResult<T> => ({
	cancelled,
	frame,
	submitted,
	submittedLabels: markedChoiceLabels(choices, marked),
	value: markedChoiceValues(choices, marked),
});

export const multipleChoicesValueResult = <T>(value: T[], submitted = false): MultipleChoicesReadResult<T> => ({
	cancelled: false,
	submitted,
	submittedLabels: [],
	value,
});

export const cancelledMultipleChoicesResult = <T>(choices: Array<Choice<T>>, marked: Set<number>, value: T[]): MultipleChoicesReadResult<T> => ({
	cancelled: true,
	submitted: false,
	submittedLabels: markedChoiceLabels(choices, marked),
	value,
});

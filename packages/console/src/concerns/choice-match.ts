import { isDeepStrictEqual } from 'node:util';
import { answerChoiceIndex } from '#console/concerns/choice-normalize';
import type { Choice } from '#console/types';

export const choiceValueEquals = (left: unknown, right: unknown): boolean => Object.is(left, right) || isDeepStrictEqual(left, right);

export const choiceByValue = <T>(choices: Array<Choice<T>>, value: unknown): Choice<T> | undefined => choices.find((choice) => !choice.disabled && choiceValueEquals(choice.value, value));

export const findChoice = <T>(choices: Array<Choice<T>>, answer: string): Choice<T> | undefined => {
	const normalizedAnswer = answer.trim();
	const index = answerChoiceIndex(normalizedAnswer);

	if (index !== null) {
		return choices[index - 1];
	}

	return choices.find((choice) => choice.label === normalizedAnswer || String(choice.value) === normalizedAnswer);
};

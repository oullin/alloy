import { parseChoiceAnswerIndex, parseChoiceRecordValue } from '#console/concerns/validators/choice-answer';
import { parseChoice, parseChoiceOptions, parseChoiceRecordEntries, parseChoiceValue } from '#console/concerns/validators/choice';
import type { Choice, ChoiceOptions } from '#console/types';

export const normalizeChoices = <T>(options: ChoiceOptions<T>): Array<Choice<T>> => {
	const parsed = parseChoiceOptions(options);

	if (parsed.kind === 'record') {
		return parseChoiceRecordEntries(parsed.options).map(({ label, value }) => ({
			label,
			value: parseChoiceRecordValue<T>(value),
		}));
	}

	return parsed.options.map((choice) => {
		const parsed = parseChoice<T>(choice);

		if (parsed) {
			return parsed;
		}

		return {
			label: String(choice),
			value: parseChoiceValue<T>(choice),
		};
	});
};

export const normalizeSearchChoices = <T>(options: ChoiceOptions<T>): Array<Choice<T>> => normalizeChoices(options);

export const answerChoiceIndex = (answer: string): number | null => parseChoiceAnswerIndex(answer.trim());

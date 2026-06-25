import { parseChoiceAnswerIndex, parseChoiceRecordValue } from '#console/concerns/validators/choice-answer';
import { parseChoice, parseChoiceOptions, parseChoiceRecordEntries, parseChoiceValue } from '#console/concerns/validators/choice';
import type { Choice, ChoiceOptions } from '#console/types';

const labelFromChoiceValue = (choice: unknown): string => {
	if (typeof choice === 'string') {
		return choice;
	}

	if (choice === undefined) {
		return 'undefined';
	}

	if (choice === null) {
		return 'null';
	}

	if (typeof choice === 'number' || typeof choice === 'boolean' || typeof choice === 'bigint' || typeof choice === 'symbol') {
		return String(choice);
	}

	if (Array.isArray(choice)) {
		return choice.map((item) => (item === null || item === undefined ? '' : labelFromChoiceValue(item))).join(',');
	}

	const customToString = Reflect.get(choice, 'toString') as unknown;

	if (typeof customToString === 'function' && customToString !== Object.prototype.toString) {
		return customToString.call(choice);
	}

	return Object.prototype.toString.call(choice);
};

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
			label: labelFromChoiceValue(choice),
			value: parseChoiceValue<T>(choice),
		};
	});
};

export const normalizeSearchChoices = <T>(options: ChoiceOptions<T>): Array<Choice<T>> => normalizeChoices(options);

export const answerChoiceIndex = (answer: string): number | null => parseChoiceAnswerIndex(answer.trim());

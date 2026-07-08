import { describe, expect, it } from 'vite-plus/test';

import { resolveTextFormArguments } from '#console/form/builder/prompts/validators/basic';
import { parseSelectChoiceOptions, parseSelectStepName } from '#console/form/builder/prompts/validators/select';
import { parseOutputScroll, parseOutputStepName } from '#console/form/builder/validators/output';
import { parseStatusCallback, parseStatusLabel } from '#console/form/builder/validators/status';
import { parseValidationResult } from '#console/validators/result';
import { requiredMessage } from '#console/validators/required';

describe('validator edge cases', () => {
	it('accepts and rejects form builder prompt argument shapes', () => {
		expect(resolveTextFormArguments('Name', 'Placeholder', 'Ada', true, undefined, 'Hint', 'name')).toEqual({
			name: 'name',
			options: {
				default: 'Ada',
				hint: 'Hint',
				label: 'Name',
				message: 'Name',
				placeholder: 'Placeholder',
				required: true,
				transform: undefined,
				validate: undefined,
			},
		});

		expect(resolveTextFormArguments({ message: 'Name' }, 'name')).toEqual({
			name: 'name',
			options: { message: 'Name' },
		});

		expect(parseSelectChoiceOptions({ red: 'Red', blue: 'Blue' })).toEqual({ red: 'Red', blue: 'Blue' });
		expect(parseSelectChoiceOptions(['Red', 'Blue'])).toEqual(['Red', 'Blue']);
		expect(() => parseSelectChoiceOptions('Red')).toThrow('Choice options must be an array or record.');
		expect(parseSelectStepName(12)).toBeUndefined();
	});

	it('validates status and output helper arguments', async () => {
		const callback = async () => 'done';

		expect(parseStatusLabel('Deploy')).toBe('Deploy');
		expect(parseStatusCallback(callback)).toBe(callback);
		expect(parseOutputStepName('summary')).toBe('summary');
		expect(parseOutputScroll(3, 10)).toBe(3);
		expect(parseStatusLabel(1)).toBeUndefined();
		expect(() => parseStatusCallback('bad')).toThrow('A status callback is required.');
		expect(parseOutputStepName(false)).toBeUndefined();
		expect(parseOutputScroll('bad', 10)).toBe(10);
	});

	it('normalizes typed validator results and required messages', () => {
		expect(parseValidationResult('Invalid.')).toBe('Invalid.');
		expect(parseValidationResult(null)).toBeNull();
		expect(parseValidationResult(undefined)).toBeUndefined();
		expect(() => parseValidationResult(false)).toThrow('The validator must return a string or null.');

		expect(requiredMessage('', true)).toBe('Required.');
		expect(requiredMessage([], 'Pick one.')).toBe('Pick one.');
		expect(requiredMessage('ok', true)).toBeUndefined();
		expect(requiredMessage('', false)).toBeUndefined();
	});
});

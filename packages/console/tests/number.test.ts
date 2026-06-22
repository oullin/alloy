import { describe, expect, it } from 'vitest';
import { createMemoryOutput, createScriptedInput, Key, number, parseAnsiText, PromptValidationError, withPromptEnvironment } from '#console/index';
import { parseNumberInput } from '#console/prompts/number/validators/value';

describe('number prompt', () => {
	it('rejects partial decimal input and accepts a retry', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['1', '2', 'a', Key.enter, '4', '2', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Amount' }),
		);

		expect(result).toBe(42);
		expect(output.text()).toContain('Must be a number');
	});

	it('casts decimal numeric values to integers', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['1', '.', '5', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count', integer: true }),
		);

		expect(result).toBe(1);
	});

	it('returns an empty string for optional empty input', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count' }),
		);

		expect(result).toBe('');
	});

	it('renders placeholders before number input', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number('Count', '0'),
		);

		expect(output.text()).toContain('\u001B[36m ┌\u001B[39m \u001B[36mCount\u001B[39m ');
		expect(parseAnsiText(output.text())).toContain('0  ▲▼');
		expect(output.text()).toContain('┌ \u001B[2mCount\u001B[22m ');
	});

	it('renders typed number values while reading raw key input', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['4', '2', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count' }),
		);

		expect(result).toBe(42);
		expect(parseAnsiText(output.text())).toContain('4  ▲▼');
		expect(parseAnsiText(output.text())).toContain('42  ▲▼');
		expect(output.text()).toContain('┌ \u001B[2mCount\u001B[22m ');
	});

	it('requires number input when configured', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter, '4', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count', required: true }),
		);

		expect(result).toBe(4);
		expect(output.text()).toContain('Required.');
	});

	it('returns decimal input as an integer value', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['1', '.', '9', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count' }),
		);

		expect(result).toBe(1);
	});

	it('applies transforms from positional number prompts', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['4', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number('Count', '', '', false, undefined, '', undefined, undefined, undefined, (value) => Number(value) * 2),
		);

		expect(result).toBe(8);
	});

	it('parses numeric values as integers', () => {
		expect(parseNumberInput('1.9')).toEqual({ value: 1 });
		expect(parseNumberInput('1.9', { integer: true })).toEqual({ value: 1 });
		expect(parseNumberInput('1', { integer: true })).toEqual({ value: 1 });
	});

	it('validates typed values against min and max bounds', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['1', Key.enter, '5', Key.enter, '3', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count', min: 2, max: 4 }),
		);

		expect(result).toBe(3);
		expect(output.text()).toContain('Must be at least 2');
		expect(output.text()).toContain('Must be less than 4');
		expect(output.text().split('┌ \u001B[2mCount\u001B[22m').length - 1).toBe(1);
	});

	it('increments and decrements with arrow keys', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.up, Key.up, Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count', step: 2 }),
		);

		expect(result).toBe(1);
		expect(parseAnsiText(output.text())).toContain('1  ▲▼');
	});

	it('uses the max value when decrementing from an empty number input', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count', max: 9 }),
		);

		expect(result).toBe(9);
		expect(parseAnsiText(output.text())).toContain('9  ▲▼');
		expect(output.text()).toContain('\u001B[2m▲\u001B[22m▼');
	});

	it('does not increment or decrement with control navigation keys', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.ctrlP, Key.ctrlP, Key.ctrlN, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count', step: 2 }),
		);

		expect(result).toBe('');
		expect(parseAnsiText(output.text())).not.toContain('1  ▲▼');
	});

	it('increments and decrements decimal values with whole steps', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.up, Key.down, Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Amount', default: 1.5, step: 0.25 }),
		);

		expect(result).toBe(0);
		expect(parseAnsiText(output.text())).toContain('2  ▲▼');
		expect(parseAnsiText(output.text())).toContain('0  ▲▼');
	});

	it('falls back to a whole step for invalid step sizes', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.up, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Amount', default: 1.5, step: 0 }),
		);

		expect(result).toBe(2);
	});

	it('renders default number values before input', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count', default: 7 }),
		);

		expect(result).toBe(7);
		expect(parseAnsiText(output.text())).toContain('7  ▲▼');
		expect(output.text()).toContain('┌ \u001B[2mCount\u001B[22m ');
	});

	it('returns an empty string for non-interactive number prompts without defaults', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				output,
				error: output,
				interactive: false,
			},
			() => number({ message: 'Count' }),
		);

		expect(result).toBe('');
	});

	it('rejects required non-interactive number prompts without defaults', async () => {
		const output = createMemoryOutput();

		await expect(
			withPromptEnvironment(
				{
					output,
					error: output,
					interactive: false,
				},
				() => number({ message: 'Count', required: true }),
			),
		).rejects.toThrow(PromptValidationError);
	});

	it('transforms non-interactive number defaults before validation and return', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				output,
				error: output,
				interactive: false,
			},
			() =>
				number({
					message: 'Count',
					default: 2,
					transform: (value) => Number(value) * 2,
					validate: (value) => (value === 4 ? null : 'Unexpected value.'),
				}),
		);

		expect(result).toBe(4);
	});

	it('renders cancelled number frames with the current value', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['4', Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count' }),
		);

		expect(result).toBe(4);
		expect(output.text()).toContain('Cancelled.');
		expect(output.text()).toContain('\u001B[31m ┌\u001B[39m Count ');
		expect(parseAnsiText(output.text())).toContain('4');
	});

	it('clamps arrow key changes to min and max', async () => {
		const output = createMemoryOutput();

		const increased = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.up, Key.up, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count', max: 3, min: 2, step: 2 }),
		);

		const decreased = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => number({ message: 'Count', max: 3, min: 2, step: 2 }),
		);

		expect(increased).toBe(3);
		expect(decreased).toBe(2);
	});
});

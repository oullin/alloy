import { afterEach, describe, expect, it } from 'vitest';
import { createMemoryOutput, withPromptEnvironment } from '#console/environment';
import { createScriptedInput } from '#console/environment/scripted-input';
import { Key } from '#console/key';
import { ask, cancelUsing, ensureRequired, promptUntilValid, PromptValidationError, validationMessage } from '#console/prompt';
import { text } from '#console/prompts/basic';

describe('prompt validation', () => {
	afterEach(() => {
		cancelUsing(null);
	});

	it('uses the default required message for empty values', () => {
		expect(ensureRequired('', true)).toBe('Required.');
		expect(ensureRequired([], true)).toBe('Required.');
		expect(ensureRequired(false, true)).toBe('Required.');
		expect(ensureRequired(null, true)).toBe('Required.');
	});

	it('accepts non-empty required arrays', () => {
		expect(ensureRequired(['choice'], true)).toBeUndefined();
	});

	it('uses custom required messages and treats empty strings as required configuration', () => {
		expect(ensureRequired('', 'Choose something.')).toBe('Choose something.');
		expect(ensureRequired('', '')).toBe('Required.');
	});

	it('ignores required validation when explicitly disabled', () => {
		expect(ensureRequired('', false)).toBeUndefined();
		expect(ensureRequired('', undefined)).toBeUndefined();
	});

	it('accepts null, undefined, and empty string validator results', async () => {
		await expect(validationMessage('value', () => null)).resolves.toBeUndefined();

		await expect(validationMessage('value', () => undefined)).resolves.toBeUndefined();

		await expect(validationMessage('value', () => '')).resolves.toBeUndefined();
	});

	it('rejects invalid validator results', async () => {
		await expect(validationMessage('value', () => false as never)).rejects.toThrow('The validator must return a string or null.');
	});

	it('retries when a prompt reader throws a validation error', async () => {
		const output = createMemoryOutput();
		const attempts: number[] = [];

		const result = await withPromptEnvironment(
			{
				output,
				error: output,
				interactive: true,
			},
			() =>
				promptUntilValid({ message: 'Name' }, async (attempt) => {
					attempts.push(attempt);

					if (attempt === 0) {
						throw new PromptValidationError('Try again.');
					}

					return 'Ada';
				}),
		);

		expect(result).toBe('Ada');
		expect(attempts).toEqual([0, 1]);
		expect(output.text()).toContain('Try again.');
	});

	it('retries required validation failures in interactive mode', async () => {
		const output = createMemoryOutput();
		const values = ['', 'Ada'];

		const result = await withPromptEnvironment(
			{
				output,
				error: output,
				interactive: true,
			},
			() => promptUntilValid({ message: 'Name', required: true }, async () => values.shift() ?? ''),
		);

		expect(result).toBe('Ada');
		expect(output.text()).toContain('Required.');
		expect(output.text()).toContain('\u001B[31m! Required.\u001B[39m');
	});

	it('rejects missing line input support', async () => {
		await withPromptEnvironment(
			{
				input: {},
				output: createMemoryOutput(),
				error: createMemoryOutput(),
				interactive: true,
			},
			async () => {
				await expect(ask('Name')).rejects.toThrow('The configured prompt input cannot read lines.');
			},
		);
	});

	it('uses a custom cancellation return value after rendering the cancelled state', async () => {
		const output = createMemoryOutput();

		cancelUsing(() => 'Manual');

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() => text('Name'),
		);

		expect(result).toBe('Manual');
		expect(output.text()).toContain('Cancelled.');
	});

	it('propagates custom cancellation errors after rendering the cancelled state', async () => {
		const output = createMemoryOutput();

		cancelUsing(() => {
			throw new Error('Stopped.');
		});

		await expect(
			withPromptEnvironment(
				{
					input: createScriptedInput([Key.ctrlC]),
					output,
					error: output,
					interactive: true,
				},
				() => text('Name'),
			),
		).rejects.toThrow('Stopped.');

		expect(output.text()).toContain('Cancelled.');
	});
});

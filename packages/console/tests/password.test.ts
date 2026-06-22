import { describe, expect, it } from 'vitest';
import { createMemoryOutput, createScriptedInput, Key, password, PromptValidationError, withPromptEnvironment } from '#console/index';

describe('password prompt', () => {
	it('returns typed values without rendering the raw password', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['s', 'e', 'c', 'r', 'e', 't', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => password('Password'),
		);

		expect(result).toBe('secret');
		expect(output.text()).not.toContain('secret');
		expect(output.text()).toContain('••••••');
		expect(output.text()).toContain('\u001B[36m ┌\u001B[39m \u001B[36mPassword\u001B[39m ');
		expect(output.text()).toContain('┌ \u001B[2mPassword\u001B[22m ');
	});

	it('renders password placeholders before input', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => password('Password', 'Required'),
		);

		expect(output.text()).toContain('\u001B[36m ┌\u001B[39m \u001B[36mPassword\u001B[39m ');
		expect(output.text()).toContain('\u001B[2mRequired\u001B[22m');
	});

	it('returns object-option defaults without rendering the raw value', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => password({ message: 'Password', default: 'secret' }),
		);

		expect(result).toBe('secret');
		expect(output.text()).not.toContain('secret');
		expect(output.text()).toContain('••••••');
	});

	it('edits object-option defaults without rendering the raw value', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.backspace, 'x', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => password({ message: 'Password', default: 'secret' }),
		);

		expect(result).toBe('secrex');
		expect(output.text()).not.toContain('secret');
		expect(output.text()).not.toContain('secrex');
		expect(output.text()).toContain('•••••');
		expect(output.text()).toContain('••••••');
	});

	it('transforms values before returning them', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['s', 'e', 'c', 'r', 'e', 't', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => password({ message: 'Password', transform: (value) => value.toUpperCase() }),
		);

		expect(result).toBe('SECRET');
		expect(output.text()).not.toContain('secret');
	});

	it('retries after validation errors without rendering the raw value', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['p', 'a', 's', Key.enter, 's', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => password({ message: 'Password', validate: (value) => (value.length < 4 ? 'Password must be at least 4 characters.' : '') }),
		);

		expect(result).toBe('pass');
		expect(output.text()).toContain('Password must be at least 4 characters.');
		expect(output.text()).not.toContain('pas');
		expect(output.text()).not.toContain('pass');
		expect(output.text().split('┌ \u001B[2mPassword\u001B[22m').length - 1).toBe(1);
	});

	it('edits values with backspace and delete keys without rendering the raw value', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['p', 'a', 'z', Key.backspace, 's', 's', Key.left, Key.delete, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => password('Password'),
		);

		expect(result).toBe('pas');
		expect(output.text()).not.toContain('pas');
		expect(output.text()).toContain('•••');
	});

	it('renders cancellation without exposing the raw value', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['s', 'e', 'c', Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() => password('Password'),
		);

		expect(result).toBe('sec');
		expect(output.text()).toContain('Cancelled.');
		expect(output.text()).toContain('\u001B[31m ┌\u001B[39m Password ');
		expect(output.text()).toContain('•••');
		expect(output.text()).not.toContain('sec');
	});

	it('returns empty strings and validates required values in non-interactive mode', async () => {
		await expect(
			withPromptEnvironment(
				{
					output: createMemoryOutput(),
					error: createMemoryOutput(),
					interactive: false,
				},
				() => password('Password'),
			),
		).resolves.toBe('');

		await expect(
			withPromptEnvironment(
				{
					output: createMemoryOutput(),
					error: createMemoryOutput(),
					interactive: false,
				},
				() => password({ message: 'Password', required: true }),
			),
		).rejects.toThrow(PromptValidationError);
	});
});

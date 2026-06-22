import { describe, expect, it } from 'vite-plus/test';
import { createMemoryOutput, createScriptedInput, Key, pause, withPromptEnvironment } from '#console/index';

describe('pause prompt', () => {
	it('waits for enter from key-driven input', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['x', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => pause('Continue'),
		);

		expect(result).toBe(true);
		expect(output.text()).toContain('\u001B[36m ┌\u001B[39m \u001B[36mContinue\u001B[39m ');
		expect(output.text().endsWith('\n')).toBe(true);
	});

	it('falls back to line input', async () => {
		const output = createMemoryOutput();

		let asked = '';

		const result = await withPromptEnvironment(
			{
				input: {
					async readLine(message: string): Promise<string> {
						asked = message;

						return '';
					},
				},
				output,
				error: output,
				interactive: true,
			},
			() => pause('Continue'),
		);

		expect(result).toBe(true);
		expect(asked).toContain('Continue');
	});

	it('returns false when the configured environment is non-interactive', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: false,
			},
			() => pause('This should not be rendered'),
		);

		expect(result).toBe(false);
		expect(output.text()).not.toContain('This should not be rendered');
	});
});

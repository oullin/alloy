import { describe, expect, it } from 'vitest';
import { autocomplete, createMemoryOutput, createScriptedInput, Key, parseAnsiText, suggest, withPromptEnvironment } from '#console/index';

const runSuggest = async (keys: string[], options: string[] | ((query: string) => string[])): Promise<{ output: string; result: string }> => {
	const output = createMemoryOutput();

	const result = await withPromptEnvironment(
		{
			input: createScriptedInput(keys),
			output,
			error: output,
			interactive: true,
		},
		() => suggest('Favorite color?', options),
	);

	return { output: output.text(), result };
};

const expectSuggestion = async (keys: string[], options: string[] | ((query: string) => string[]), result: string): Promise<void> => {
	await expect(runSuggest(keys, options)).resolves.toMatchObject({ result });
};

const outputFor = async (keys: string[], options: string[] | ((query: string) => string[])): Promise<string> => {
	const output = createMemoryOutput();

	await withPromptEnvironment(
		{
			input: createScriptedInput(keys),
			output,
			error: output,
			interactive: true,
		},
		() => suggest('Favorite color?', options),
	);

	return output.text();
};

describe('suggest prompt', () => {
	it('accepts arbitrary typed input', async () => {
		await expectSuggestion(['B', 'l', 'a', 'c', 'k', Key.enter], ['Red', 'Green', 'Blue'], 'Black');
	});

	it('completes input using tab', async () => {
		await expectSuggestion(['b', Key.tab, Key.enter], ['Red', 'Green', 'Blue'], 'Blue');
	});

	it('navigates matches using arrow and emacs keys', async () => {
		await expectSuggestion(['b', Key.down, Key.down, Key.down, Key.up, Key.enter], ['Red', 'Blue', 'Black', 'Blurple'], 'Black');

		await expectSuggestion(['b', Key.ctrlN, Key.ctrlN, Key.ctrlN, Key.ctrlP, Key.enter], ['Red', 'Blue', 'Black', 'Blurple'], 'Black');
	});

	it('navigates suggestions backwards with shift tab', async () => {
		await expectSuggestion(['b', Key.shiftTab, Key.enter], ['Red', 'Blue', 'Black', 'Blurple'], 'Blurple');
	});

	it('supports page suggestion navigation keys', async () => {
		await expectSuggestion(['b', Key.down, Key.pageDown, Key.pageUp, Key.pageDown, Key.enter], ['Red', 'Blue', 'Black', 'Blurple'], 'Blurple');
	});

	it('supports home and end suggestion navigation keys', async () => {
		await expectSuggestion(['b', Key.down, Key.end[0], Key.home[0], Key.enter], ['Red', 'Blue', 'Black', 'Blurple'], 'Blue');
	});

	it('supports control-line suggestion navigation keys', async () => {
		await expectSuggestion(['b', Key.down, Key.ctrlE, Key.enter], ['Red', 'Blue', 'Black', 'Blurple'], 'Blurple');

		await expectSuggestion(['b', Key.down, Key.ctrlE, Key.ctrlA, Key.enter], ['Red', 'Blue', 'Black', 'Blurple'], 'Blue');
	});

	it('supports callback options', async () => {
		await expectSuggestion(['e', 'e', Key.down, Key.enter], (value) => ['Red', 'Green', 'Blue'].filter((option) => option.toLowerCase().includes(value.toLowerCase())), 'Green');
	});

	it('renders typed queries and suggestion matches', async () => {
		const output = await outputFor(['b', Key.down, Key.enter], ['Red', 'Green', 'Blue']);

		expect(output).toContain('\u001B[36m ┌\u001B[39m \u001B[36mFavorite color?\u001B[39m ');
		expect(parseAnsiText(output)).toContain('b');
		expect(output).toContain('›');
		expect(output).toContain('Blue');
		expect(output).toContain('┌ \u001B[2mFavorite color?\u001B[22m ');
	});

	it('highlights suggestions on tab without replacing typed input', async () => {
		const output = await outputFor(['b', Key.tab, Key.enter], ['Red', 'Green', 'Blue']);

		expect(parseAnsiText(output)).toContain('b');
		expect(output).not.toContain('Favorite color? Blue');
	});

	it('clears highlighted suggestions with horizontal navigation keys', async () => {
		await expectSuggestion(['b', Key.down, Key.left, Key.enter], ['Red', 'Green', 'Blue'], 'b');
	});

	it('renders suggest info for the highlighted result', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => suggest({ message: 'Favorite color?', options: ['Red', 'Green', 'Blue'], info: (value) => `About ${value ?? 'none'}` }),
		);

		expect(output.text()).toContain('About Blue');
	});

	it('supports label-first suggest helpers with info', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => suggest('Favorite color?', ['Red', 'Green', 'Blue'], '', '', 5, false, undefined, '', undefined, (value) => `About ${value ?? 'none'}`),
		);

		expect(output.text()).toContain('About Blue');
	});

	it('renders cancelled suggest frames with the current value', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() => suggest('Favorite color?', ['Blue']),
		);

		expect(result).toBe('b');
		expect(output.text()).toContain('Cancelled.');
		expect(output.text()).toContain('\u001B[31m ┌\u001B[39m Favorite color? ');
		expect(parseAnsiText(output.text())).toContain('b');
	});
});

describe('autocomplete prompt', () => {
	it('renders autocomplete prompts with question formatting, hints, and placeholders', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Red', 'Green', 'Blue'], 'Type a color', '', false, undefined, 'Optional'),
		);

		expect(output.text()).toContain('\u001B[36m ┌\u001B[39m \u001B[36mFavorite color?\u001B[39m ');
		expect(output.text()).toContain('\u001B[2mType a color\u001B[22m');
		expect(output.text()).toContain('┌ \u001B[2mFavorite color?\u001B[22m ');
	});

	it('renders autocomplete ghost text without a suggestion list', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Red', 'Green', 'Blue']),
		);

		expect(parseAnsiText(output.text())).toContain('blue');
		expect(output.text()).toContain('b\u001B[7ml\u001B[27m\u001B[2mue\u001B[22m');
		expect(output.text()).not.toContain('›');
	});

	it('supports label-first autocomplete helpers with info', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Red', 'Green', 'Blue'], '', '', false, undefined, '', undefined, (value) => `About ${value ?? 'none'}`),
		);

		expect(output.text()).toContain('About Blue');
	});

	it('accepts ghost completion with tab', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.tab, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Red', 'Green', 'Blue']),
		);

		expect(result).toBe('Blue');
		expect(output.text()).toContain('┌ \u001B[2mFavorite color?\u001B[22m ');
		expect(parseAnsiText(output.text())).toContain('Blue');
	});

	it('does not accept tab completion before the cursor reaches the end', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.left, Key.tab, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Blue']),
		);

		expect(result).toBe('b');
	});

	it('hides autocomplete ghost text while editing before the end', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.left, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Blue']),
		);

		const frames = output.text().trimEnd().split('\n');

		expect(parseAnsiText(output.text())).toContain('b');
		expect(frames.at(-1)).not.toContain('Blue');
	});

	it('accepts same-length autocomplete matches with right arrow', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', 'l', 'u', 'e', Key.right, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Blue']),
		);

		expect(result).toBe('Blue');
	});

	it('cycles autocomplete matches before accepting completion', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.down, Key.tab, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Blue', 'Black', 'Blurple']),
		);

		expect(result).toBe('Black');
	});

	it('does not cycle autocomplete matches with control navigation keys', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.ctrlN, Key.tab, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Blue', 'Black', 'Blurple']),
		);

		expect(result).toBe('Blue');
	});

	it('does not cycle autocomplete matches with shift tab', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.shiftTab, Key.tab, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Blue', 'Black', 'Blurple']),
		);

		expect(result).toBe('Blue');
	});

	it('does not cycle autocomplete matches with page navigation keys', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.pageDown, Key.pageUp, Key.pageDown, Key.tab, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete({ message: 'Favorite color?', options: ['Blue', 'Black', 'Blurple'], scroll: 2 }),
		);

		expect(result).toBe('Blue');
		expect(output.text()).toContain('┌ \u001B[2mFavorite color?\u001B[22m ');
		expect(parseAnsiText(output.text())).toContain('Blue');
	});

	it('renders cancelled autocomplete frames with the current value', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['b', Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() => autocomplete('Favorite color?', ['Blue']),
		);

		expect(result).toBe('b');
		expect(output.text()).toContain('Cancelled.');
		expect(output.text()).toContain('\u001B[31m ┌\u001B[39m Favorite color? ');
		expect(parseAnsiText(output.text())).toContain('b');
	});
});

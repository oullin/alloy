import { describe, expect, it } from 'vite-plus/test';
import { createMemoryOutput, createScriptedInput, Key, multisearch, parseAnsiText, PromptValidationError, search, withPromptEnvironment } from '#console/index';
import type { SearchPromptOptions } from '#console/index';

const colors = (value: string): Record<string, string> => {
	const options = {
		red: 'Red',
		green: 'Green',
		blue: 'Blue',
	};

	return Object.fromEntries(Object.entries(options).filter(([, label]) => label.toLowerCase().includes(value.toLowerCase())));
};

describe('search prompt', () => {
	it('returns keys from associative options', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['u', 'e', Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors }),
		);

		expect(result).toBe('blue');
		expect(output.text()).toContain('\u001B[36m ┌\u001B[39m \u001B[36mFavorite color?\u001B[39m ');
		expect(parseAnsiText(output.text())).toContain('│ ue');
		expect(output.text()).toContain('┌ \u001B[2mFavorite color?\u001B[22m ');
		expect(output.text()).toContain('│ Blue');
	});

	it('validates selected values and allows another selection', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter, Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				search({
					message: 'Favorite color?',
					options: () => ({ red: 'Red', green: 'Green', blue: 'Blue' }),
					validate: (value) => (value === 'red' ? 'Please choose green.' : null),
				}),
		);

		expect(result).toBe('green');
		expect(output.text()).toContain('Please choose green.');
		expect(parseAnsiText(output.text())).not.toMatch(/^ │ Red\s*│$/m);
		expect(parseAnsiText(output.text())).toMatch(/^ │ Green\s*│$/m);
	});

	it('renders the typed query while navigating results', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput(['r', Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors }),
		);

		expect(parseAnsiText(output.text())).toContain('│ r');
		expect(output.text()).toContain('\u001B[36m›\u001B[39m Red');
		expect(output.text()).toContain('\u001B[2mGreen\u001B[22m');
	});

	it('does not render an active search row before navigation highlights a result', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput([Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors, default: 'red' }),
		);

		expect(output.text()).not.toContain('›');
		expect(output.text()).toContain('\u001B[2mRed\u001B[22m');
	});

	it('respects scroll windows when rendering results', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.down, Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				search({
					message: 'Pick number',
					options: ['one', 'two', 'three', 'four'],
					scroll: 3,
				}),
		);

		expect(result).toBe('three');
		expect(parseAnsiText(output.text())).toContain('│   two');
		expect(parseAnsiText(output.text())).toContain('│ › three');
		expect(parseAnsiText(output.text())).toContain('│   four');
		expect(output.text()).toContain('┃');
		expect(output.text()).toContain('│');
	});

	it('supports page search navigation keys', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.pageDown, Key.pageUp, Key.pageDown, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				search({
					message: 'Pick number',
					options: ['one', 'two', 'three', 'four'],
					scroll: 2,
				}),
		);

		expect(result).toBe('three');
		expect(output.text()).toContain('│ three');
	});

	it('supports control-key search navigation', async () => {
		const output = createMemoryOutput();

		const forwardResult = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.ctrlN, Key.ctrlN, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors }),
		);

		const backwardResult = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.ctrlP, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors }),
		);

		expect(forwardResult).toBe('green');
		expect(backwardResult).toBe('blue');
	});

	it('renders search info for the highlighted result', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors, info: (value) => `About ${value ?? 'none'}` }),
		);

		expect(output.text()).toContain('About red');
	});

	it('supports label-first search helpers with info', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => search('Favorite color?', colors, '', 5, undefined, '', true, undefined, (value) => `About ${value ?? 'none'}`),
		);

		expect(result).toBe('red');
		expect(output.text()).toContain('About red');
	});

	it('renders an empty result message for unmatched search queries', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput(['z', Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors, default: 'red' }),
		);

		expect(parseAnsiText(output.text())).toContain('No results.');
		expect(output.text()).toContain('\u001B[2m  No results.\u001B[22m');
	});

	it('skips disabled search results while navigating', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				search({
					message: 'Favorite color?',
					options: [
						{ label: 'Red', value: 'red', disabled: true },
						{ label: 'Green', value: 'green' },
					],
				}),
		);

		expect(result).toBe('green');
	});

	it('keeps search highlights empty when submitting without a highlighted result', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter, Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors }),
		);

		expect(result).toBe('red');
	});

	it('clears search highlights with horizontal navigation keys', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.left, Key.enter, Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors }),
		);

		expect(result).toBe('red');
	});

	it('cancels search prompts with the current highlighted result', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors }),
		);

		expect(result).toBe('red');
		expect(output.text()).toContain('Cancelled.');
		expect(output.text()).toContain('\u001B[31m ┌\u001B[39m \u001B[2mFavorite color?\u001B[22m ');
		expect(parseAnsiText(output.text())).toContain('⚠ Cancelled.');
	});

	it('uses the default when cancelling a disabled highlighted search result', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				search({
					message: 'Favorite color?',
					options: [
						{ label: 'Red', value: 'red', disabled: true },
						{ label: 'Green', value: 'green' },
					],
					default: 'green',
				}),
		);

		expect(result).toBe('green');
		expect(output.text()).toContain('Cancelled.');
	});

	it('rejects optional single search prompts', async () => {
		const options = {
			message: 'Favorite color?',
			options: colors,
			required: false,
		} as unknown as SearchPromptOptions<string>;

		await expect(search(options)).rejects.toThrow('Argument [required] must be true or a string.');
	});

	it('transforms search values before validation and return', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				search({
					message: 'Favorite color?',
					options: colors,
					transform: (value) => value.toUpperCase(),
					validate: (value) => (value === 'RED' ? null : 'Unexpected value.'),
				}),
		);

		expect(result).toBe('RED');
	});

	it('returns search defaults for empty line-mode answers', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: {
					async readLine(): Promise<string> {
						return '';
					},
				},
				output,
				error: output,
				interactive: true,
			},
			() =>
				search({
					message: 'Favorite color?',
					options: colors,
					default: 'green',
				}),
		);

		expect(result).toBe('green');
	});

	it('returns search defaults for empty key-mode submissions', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				search({
					message: 'Favorite color?',
					options: colors,
					default: 'green',
				}),
		);

		expect(result).toBe('green');
		expect(parseAnsiText(output.text())).toMatch(/^ │ Green\s*│$/m);
	});

	it('transforms non-interactive search defaults before validation and return', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				output,
				error: output,
				interactive: false,
			},
			() =>
				search({
					message: 'Favorite color?',
					options: colors,
					default: 'green',
					transform: (value) => value.toUpperCase(),
					validate: (value) => (value === 'GREEN' ? null : 'Unexpected value.'),
				}),
		);

		expect(result).toBe('GREEN');
	});

	it('trims line-mode search answers before matching choices', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: {
					async readLine(): Promise<string> {
						return ' Green ';
					},
				},
				output,
				error: output,
				interactive: true,
			},
			() => search({ message: 'Favorite color?', options: colors }),
		);

		expect(result).toBe('green');
	});

	it('rejects disabled exact line-mode matches and retries', async () => {
		const output = createMemoryOutput();
		const answers = ['Red', 'Green'];

		const result = await withPromptEnvironment(
			{
				input: {
					async readLine(): Promise<string> {
						return answers.shift() ?? '';
					},
				},
				output,
				error: output,
				interactive: true,
			},
			() =>
				search({
					message: 'Favorite color?',
					options: [
						{ label: 'Red', value: 'red', disabled: true },
						{ label: 'Green', value: 'green' },
					],
				}),
		);

		expect(result).toBe('green');
		expect(output.text()).toContain('Please select a valid option.');
	});
});

describe('multisearch prompt', () => {
	it('toggles highlighted search results', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['V', Key.down, Key.space, Key.backspace, 'G', 'r', Key.down, Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				multisearch({
					message: 'Favorite colors?',
					options: (value) => {
						const options = { green: 'Green', violet: 'Violet' };

						return Object.fromEntries(Object.entries(options).filter(([, label]) => label.toLowerCase().includes(value.toLowerCase())));
					},
				}),
		);

		expect(result).toEqual(['violet', 'green']);
		expect(output.text()).toContain('\u001B[36m› ◼\u001B[39m Violet');
		expect(output.text()).toContain('1 selected');
		expect(output.text()).toContain('2 selected');
		expect(output.text()).toContain('┌ \u001B[2mFavorite colors?\u001B[22m ');
		expect(output.text()).toContain('│ Violet');
		expect(output.text()).toContain('│ Green');
	});

	it('starts multisearch prompts with default selected values', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				multisearch({
					message: 'Favorite colors?',
					options: colors,
					default: ['green'],
				}),
		);

		expect(result).toEqual(['green']);
		expect(output.text()).toContain('1 selected');
		expect(output.text()).toContain('│ Green');
	});

	it('returns selected defaults when multisearch input is exhausted', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				multisearch({
					message: 'Favorite colors?',
					options: colors,
					default: ['green'],
				}),
		);

		expect(result).toEqual(['green']);
		expect(output.text()).toContain('│ Green');
	});

	it('returns an empty array for non-interactive multisearch prompts without defaults', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				output,
				error: output,
				interactive: false,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors }),
		);

		expect(result).toEqual([]);
	});

	it('transforms non-interactive multisearch defaults before validation and return', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				output,
				error: output,
				interactive: false,
			},
			() =>
				multisearch({
					message: 'Favorite colors?',
					options: colors,
					default: ['green'],
					transform: (value) => [...value, 'red'],
					validate: (value) => (value.length === 2 ? null : 'Unexpected value.'),
				}),
		);

		expect(result).toEqual(['green', 'red']);
	});

	it('rejects disabled line-mode multisearch choices and retries', async () => {
		const output = createMemoryOutput();
		const answers = ['Red', 'Green'];

		const result = await withPromptEnvironment(
			{
				input: {
					async readLine(): Promise<string> {
						return answers.shift() ?? '';
					},
				},
				output,
				error: output,
				interactive: true,
			},
			() =>
				multisearch({
					message: 'Favorite colors?',
					options: [
						{ label: 'Red', value: 'red', disabled: true },
						{ label: 'Green', value: 'green' },
					],
				}),
		);

		expect(result).toEqual(['green']);
		expect(output.text()).toContain('Please select valid options.');
	});

	it('rejects required non-interactive multisearch prompts without defaults', async () => {
		const output = createMemoryOutput();

		await expect(
			withPromptEnvironment(
				{
					output,
					error: output,
					interactive: false,
				},
				() => multisearch({ message: 'Favorite colors?', options: colors, required: true }),
			),
		).rejects.toThrow(PromptValidationError);
	});

	it('renders multisearch submitted frames only after validation passes', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter, Key.down, Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors, required: true }),
		);

		const rendered = parseAnsiText(output.text());

		expect(result).toEqual(['red']);
		expect(rendered).toContain('Required.');
		expect(rendered).not.toMatch(/^ │ None\s*│$/m);
		expect(rendered).toMatch(/^ │ Red\s*│$/m);
	});

	it('renders multisearch info for the highlighted result', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors, info: (value) => `About ${value ?? 'none'}` }),
		);

		expect(output.text()).toContain('About red');
		expect(output.text()).toContain('0 selected');
	});

	it('supports label-first multisearch helpers with info', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch('Favorite colors?', colors, '', 5, false, undefined, 'Use the space bar to select options.', undefined, (value) => `About ${value ?? 'none'}`),
		);

		expect(result).toEqual(['red']);
		expect(output.text()).toContain('About red');
		expect(output.text()).toContain('1 selected');
	});

	it('renders an empty result message for unmatched multisearch queries', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput(['z', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors }),
		);

		expect(parseAnsiText(output.text())).toContain('No results.');
		expect(output.text()).toContain('\u001B[2m  No results.\u001B[22m');
	});

	it('renders hidden selected counts when multisearch results are filtered', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.space, 'B', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors }),
		);

		expect(output.text()).toContain('1 selected (1 hidden)');
	});

	it('keeps default multisearch labels while filtered out', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['B', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				multisearch({
					message: 'Favorite colors?',
					options: colors,
					default: ['green'],
				}),
		);

		expect(result).toEqual(['green']);
		expect(output.text()).toContain('1 selected (1 hidden)');
		expect(output.text()).toContain('│ Green');
	});

	it('cancels multisearch prompts with the current selected values', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.space, 'B', Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors }),
		);

		expect(result).toEqual(['red']);
		expect(output.text()).toContain('Cancelled.');
		expect(output.text()).toContain('\u001B[31m ┌\u001B[39m \u001B[2mFavorite colors?\u001B[22m ');
		expect(output.text()).toContain('\u001B[9m\u001B[2mB\u001B[22m\u001B[29m');
		expect(parseAnsiText(output.text())).toContain('│ B');
	});

	it('toggles all current multisearch results with ctrl-a', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.ctrlA, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors }),
		);

		expect(result).toEqual(['red', 'green', 'blue']);
		expect(output.text()).toContain('3 selected');
		expect(output.text()).toContain('│ Red');
		expect(output.text()).toContain('│ Green');
		expect(output.text()).toContain('│ Blue');
	});

	it('does not navigate multisearch results with ctrl-n or ctrl-p', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.ctrlN, Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors }),
		);

		expect(result).toEqual([]);
		expect(output.text()).not.toContain('Selected: Green');
	});

	it('supports tab and reverse-tab multisearch navigation', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.tab, Key.tab, Key.shiftTab, Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors }),
		);

		expect(result).toEqual(['red']);
	});

	it('keeps selected multisearch values visible after clearing the query', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['V', Key.down, Key.space, Key.backspace, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				multisearch({
					message: 'Favorite colors?',
					options: (value): Record<string, string> => {
						if (value === '') {
							return { green: 'Green' };
						}

						return { violet: 'Violet' };
					},
				}),
		);

		const latestFrame = output.text().split('Favorite colors?\n').at(-1) ?? '';

		expect(result).toEqual(['violet']);
		expect(latestFrame).toContain('Violet');
		expect(latestFrame).toContain('Green');
	});

	it('supports home and end multisearch navigation keys', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.end[0], Key.space, Key.home[0], Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors }),
		);

		expect(result).toEqual(['blue', 'red']);
		expect(output.text()).toContain('2 selected');
		expect(output.text()).toContain('│ Blue');
		expect(output.text()).toContain('│ Red');
	});

	it('ignores end-of-line control navigation while multisearch results are highlighted', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.ctrlE, Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors }),
		);

		expect(result).toEqual(['red']);
		expect(output.text()).toContain('1 selected');
		expect(output.text()).toContain('│ Red');
	});

	it('supports page multisearch navigation keys', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.pageDown, Key.space, Key.pageUp, Key.space, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors, scroll: 2 }),
		);

		expect(result).toEqual(['blue', 'red']);
		expect(output.text()).toContain('2 selected');
		expect(output.text()).toContain('│ Blue');
		expect(output.text()).toContain('│ Red');
		expect(output.text()).toContain('┃');
		expect(output.text()).toContain('│');
	});

	it('skips disabled multisearch results while navigating and toggling all', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.ctrlA, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				multisearch({
					message: 'Favorite colors?',
					options: [
						{ label: 'Red', value: 'red', disabled: true },
						{ label: 'Green', value: 'green' },
						{ label: 'Blue', value: 'blue' },
					],
				}),
		);

		expect(result).toEqual(['green', 'blue']);
		expect(output.text()).toContain('2 selected');
		expect(output.text()).toContain('│ Green');
		expect(output.text()).toContain('│ Blue');
	});

	it('transforms multisearch values before validation and return', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.ctrlA, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				multisearch({
					message: 'Favorite colors?',
					options: colors,
					transform: (value) => value.toReversed(),
					validate: (value) => (value[0] === 'blue' ? null : 'Unexpected order.'),
				}),
		);

		expect(result).toEqual(['blue', 'green', 'red']);
	});

	it('trims line-mode multisearch answers before matching choices', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: {
					async readLine(): Promise<string> {
						return ' Green ';
					},
				},
				output,
				error: output,
				interactive: true,
			},
			() => multisearch({ message: 'Favorite colors?', options: colors }),
		);

		expect(result).toEqual(['green']);
	});
});

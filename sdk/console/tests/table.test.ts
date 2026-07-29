import { describe, expect, it } from 'vite-plus/test';
import { createMemoryOutput, createScriptedInput, datatable, form, Key, parseAnsiText, table, visibleWidth, withPromptEnvironment } from '#console/index';

describe('table output', () => {
	it('renders static table rows', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			table(['Name', 'Runtime'], [['Ollin', 'OpenTUI']]);
		});

		expect(output.text()).toContain('| Name  | Runtime |');
		expect(output.text()).toContain('| Ollin | OpenTUI |');
	});

	it('renders object-form table options with inferred headers', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			table({
				rows: [
					{ Name: 'Ollin', Runtime: 'OpenTUI' },
					{ Name: 'Prompts', Runtime: 'TypeScript' },
				],
			});
		});

		expect(output.text()).toContain('| Name    | Runtime    |');
		expect(output.text()).toContain('| Prompts | TypeScript |');
	});

	it('pads ANSI-styled table cells by visible width', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			table(
				['Name', 'Runtime'],
				[
					['\u001B[31mRed\u001B[39m', 'Node'],
					['Blue', 'OpenTUI'],
				],
			);
		});

		expect(output.text()).toContain('\u001B[31mRed\u001B[39m  | Node');
		expect(output.text()).toContain('| Blue | OpenTUI |');
	});

	it('pads wide Unicode table cells by visible width', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			table(
				['Name', 'Runtime'],
				[
					['東京', 'Node'],
					['Paris', 'OpenTUI'],
				],
			);
		});

		expect(output.text()).toContain('| 東京  | Node');
		expect(output.text()).toContain('| Paris | OpenTUI |');
	});

	it('renders missing table cells as empty columns', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			table(['Name', 'Runtime'], [['Ollin'], ['Prompts', 'TypeScript']]);
		});

		expect(output.text()).toContain('| Ollin   |            |');
		expect(output.text()).toContain('| Prompts | TypeScript |');
	});

	it('renders object table rows in explicit header order', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			table({
				headers: ['Runtime', 'Name', 'Notes'],
				rows: [
					{ Name: 'Ollin', Runtime: 'OpenTUI' },
					{ Name: 'Prompts', Runtime: 'TypeScript', Notes: 'Port' },
				],
			});
		});

		expect(output.text()).toContain('| Runtime    | Name    | Notes |');
		expect(output.text()).toContain('| OpenTUI    | Ollin   |       |');
		expect(output.text()).toContain('| TypeScript | Prompts | Port  |');
	});

	it('normalizes table row cells through the validator layer', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			table(['Name', 'Enabled', 'Notes'], [{ Enabled: false, Name: 'Ollin', Notes: null }]);
		});

		expect(output.text()).toContain('| Name  | Enabled | Notes |');
		expect(output.text()).toContain('| Ollin | false   |       |');
	});

	it('renders tables without headers from row-only input', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			table([
				['Ollin', 'OpenTUI'],
				['Prompts', 'TypeScript'],
			]);
		});

		expect(output.text()).not.toContain('---');
		expect(output.text()).toContain('| Ollin   | OpenTUI    |');
		expect(output.text()).toContain('| Prompts | TypeScript |');
	});

	it('renders option-object tables with array rows as headerless tables unless headers are explicit', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			table({
				rows: [
					['Ollin', 'OpenTUI'],
					['Prompts', 'TypeScript'],
				],
			});
		});

		expect(output.text()).not.toContain('| 0       | 1          |');
		expect(output.text()).not.toContain('| ------- | ---------- |');
		expect(output.text()).toContain('| Ollin   | OpenTUI    |');
		expect(output.text()).toContain('| Prompts | TypeScript |');
	});

	it('rejects invalid table runtime shapes through the validator layer', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			expect(() => table(null as never)).toThrow();
			expect(() => table(['Name'], null as never)).not.toThrow();
			expect(() => table(['Name'], [null] as never)).toThrow();
		});
	});
});

describe('data table prompt', () => {
	it('returns the selected explicit row value', async () => {
		const output = createMemoryOutput();
		const selected = { id: 2 };

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					headers: ['Name', 'Runtime'],
					rows: [
						{ value: { id: 1 }, cells: { Name: 'Alpha', Runtime: 'Node' } },
						{ value: selected, cells: { Name: 'Beta', Runtime: 'OpenTUI' } },
					],
				}),
		);

		expect(result).toBe(selected);
		expect(output.text()).toContain('Pick project');
		expect(output.text()).toContain('›');
	});

	it('supports label-first data table helpers with filters', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['/', 'slow', Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable(
					['Name', 'Tag'],
					[
						{ Name: 'Alpha', Tag: 'fast', value: 'alpha' },
						{ Name: 'Beta', Tag: 'slow', value: 'beta' },
					],
					10,
					'Pick project',
					'',
					false,
					undefined,
					undefined,
					(query, row: any) => !Array.isArray(row) && row.Tag === query,
				),
		);

		expect(result).toBe('beta');
		expect(output.text()).toContain('Pick project slow');
		expect(output.text()).toContain('Beta');
	});

	it('transforms selected data table values before return', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					rows: [{ Name: 'Alpha', value: 'alpha' }],
					transform: (value) => String(value).toUpperCase(),
				}),
		);

		expect(result).toBe('ALPHA');
	});

	it('rejects invalid data table runtime shapes through the validator layer', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			await expect(datatable(null as never)).rejects.toThrow();

			await expect(datatable(['Name'], [null] as never)).rejects.toThrow();
		});
	});

	it('renders data table submitted frames only after validation passes', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter, Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
					validate: (value) => (value === 'beta' ? null : 'Choose beta.'),
				}),
		);

		const rendered = parseAnsiText(output.text());

		expect(result).toBe('beta');
		expect(rendered).toContain('Choose beta.');
		expect(rendered).not.toMatch(/^Alpha$/m);
		expect(rendered).toMatch(/^Beta$/m);
	});

	it('starts data table selection on matching default values', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					default: 'beta',
					message: 'Pick project',
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('beta');
		expect(output.text()).toContain('Beta');
	});

	it('returns the current data table selection when key input is exhausted', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					default: 'beta',
					message: 'Pick project',
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('beta');
		expect(output.text()).toContain('Beta');
	});

	it('filters rows in explicit search mode', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['/', 'B', Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('beta');
		expect(output.text()).toContain('Pick project B');
	});

	it('uses custom data table filters for search mode', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['/', 'slow', Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					filter: (query, row: any) => !Array.isArray(row) && row.Tag === query,
					rows: [
						{ Name: 'Alpha', Tag: 'fast', value: 'alpha' },
						{ Name: 'Beta', Tag: 'slow', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('beta');
		expect(output.text()).toContain('Pick project slow');
	});

	it('does not filter rows from printable keys in browse mode', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['B', Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('alpha');
	});

	it('cancels data table search with escape', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['/', 'B', Key.escape, Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('beta');

		const latestFrame = output.text().split('Pick project\n').at(-2) ?? '';

		expect(latestFrame).not.toContain('Pick project B');
		expect(latestFrame).toContain('Alpha');
		expect(latestFrame).toContain('Beta');
	});

	it('renders a no-results row for unmatched data table searches', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['/', 'Z', Key.escape, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					headers: ['Name'],
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('alpha');
		expect(output.text()).toContain('Pick project Z');
		expect(output.text()).toContain('No results found.');
	});

	it('cancels data table prompts with the current selected row', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('beta');
		expect(output.text()).toContain('Cancelled.');
		expect(parseAnsiText(output.text())).toContain('⚠ Cancelled.');
		expect(output.text()).toContain('/ Search');
	});

	it('cancels data table search with the current filtered row', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['/', 'B', Key.ctrlC]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('beta');
		expect(output.text()).toContain('Cancelled.');
		expect(parseAnsiText(output.text())).toContain('⚠ Cancelled.');
	});

	it('renders a submitted data table row summary', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					headers: ['Name', 'Role'],
					rows: [
						['Alice', 'Designer'],
						['Bob', 'Developer'],
					],
				}),
		);

		expect(result).toBe(1);
		expect(output.text()).toContain('Bob, Developer');
	});

	it('renders multiline data table cells as visual table rows', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name', 'Role'],
					rows: [
						['Alice', 'Lead\nDeveloper'],
						['Bob', 'Designer'],
					],
				}),
		);

		const activeFrame = output.text().split('Pick row\n').at(-2) ?? '';

		expect(result).toBe(0);
		expect(activeFrame).toContain('| › | Alice | Lead      |');
		expect(activeFrame).toContain('|   |       | Developer |');
		expect(activeFrame).toContain('|   | Bob   | Designer  |');
	});

	it('keeps highlighted multiline data table rows fully visible', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name', 'Role'],
					rows: [
						['Alice', 'Designer'],
						['Bob', 'CEO\nCTO\nDeveloper'],
						['Charlie', 'Designer'],
					],
					scroll: 5,
				}),
		);

		const activeFrame = output.text().split('Pick row\n').at(-2) ?? '';

		expect(result).toBe(1);
		expect(activeFrame).toContain('CEO');
		expect(activeFrame).toContain('CTO');
		expect(activeFrame).toContain('Developer');
	});

	it('maintains fixed data table visual height', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['Alice'], ['Bob']],
					scroll: 5,
				}),
		);

		const activeFrame = output.text().split('Pick row\n').at(-2) ?? '';

		const tableLines = activeFrame.split('\n').filter((line) => line.startsWith('|'));

		expect(result).toBe(0);
		expect(tableLines).toHaveLength(7);
		expect(tableLines.slice(2)).toHaveLength(5);
	});

	it('uses comfortable data table width for compact rows', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['A', 'B'],
					rows: [['Hi', 'Lo']],
					scroll: 5,
				}),
		);

		const activeFrame = output.text().split('Pick row\n').at(-2) ?? '';

		const widestLine = Math.max(...activeFrame.split('\n').map(visibleWidth));

		expect(result).toBe(0);
		expect(widestLine).toBeLessThan(70);
	});

	it('truncates outlier data table columns before they stretch the table', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.end[0], Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name', 'Value'],
					rows: [
						['Alice', 'Short'],
						['Bob', 'Short'],
						['Charlie', 'Short'],
						['Diana', 'Short'],
						['Ethan', 'Short'],
						['An extremely long value that should be treated as an outlier and truncated', 'Short'],
					],
					scroll: 5,
				}),
		);

		const activeFrame = output.text().split('Pick row\n').at(-2) ?? '';

		const widestLine = Math.max(...activeFrame.split('\n').map(visibleWidth));

		expect(result).toBe(5);
		expect(widestLine).toBeLessThan(76);
		expect(activeFrame).toContain('...');
	});

	it('renders data table rows without headers', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: [],
					rows: [
						['Alice', 'Developer'],
						['Bob', 'Designer'],
					],
				}),
		);

		const activeFrame = output.text().split('Pick row\n').at(-2) ?? '';

		expect(result).toBe(0);
		expect(activeFrame).not.toContain('---');
		expect(activeFrame).toContain('| › | Alice | Developer |');
		expect(activeFrame).toContain('|   | Bob   | Designer  |');
	});

	it('renders blank data table cells as empty columns', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name', 'Role'],
					rows: [
						['Alice', null],
						['Bob', undefined],
					],
				}),
		);

		const activeFrame = output.text().split('Pick row\n').at(-2) ?? '';

		expect(result).toBe(0);
		expect(activeFrame).toContain('| Name  | Role |');
		expect(activeFrame).toMatch(/\| › \| Alice \|\s+\|/);
		expect(activeFrame).toMatch(/\|   \| Bob\s+\|\s+\|/);
	});

	it('edits Unicode data table search queries without corrupting input', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['/', '😀', Key.backspace, 'B', Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('beta');
		expect(output.text()).toContain('Pick project B');
	});

	it('supports cursor edits inside data table search queries', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput(['/', 'B', 't', Key.left, 'e', Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick project',
					rows: [
						{ Name: 'Alpha', value: 'alpha' },
						{ Name: 'Beta', value: 'beta' },
					],
				}),
		);

		expect(result).toBe('beta');
		expect(output.text()).toContain('Pick project Bet');
	});

	it('falls back to row indexes when rows have no explicit value', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second']],
				}),
		);

		expect(result).toBe(1);
	});

	it('uses the first visible row without keyboard input support', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: {},
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [
						{ Name: 'First', value: 'first' },
						{ Name: 'Second', value: 'second' },
					],
				}),
		);

		expect(result).toBe('first');
	});

	it('renders a scrolling data table window around the selected row', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second'], ['Third'], ['Fourth']],
					scroll: 3,
				}),
		);

		const latestFrame = output.text().split('Pick row\n').at(-2) ?? '';

		expect(result).toBe(2);
		expect(latestFrame).not.toContain('First');
		expect(latestFrame).toContain('Second');
		expect(latestFrame).toContain('Third');
		expect(latestFrame).toContain('Fourth');
		expect(latestFrame).toContain('Viewing 2-4 of 4');
		expect(latestFrame).toContain('┃');
	});

	it('does not render data table viewing info when all rows are visible', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput([Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second']],
					scroll: 5,
				}),
		);

		expect(output.text()).not.toContain('Viewing');
		expect(output.text()).not.toContain('┃');
	});

	it('renders data table viewing info for filtered result windows', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment(
			{
				input: createScriptedInput(['/', 'a', Key.pageDown, Key.enter, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['Alpha'], ['Atlas'], ['Beta'], ['Gamma'], ['Delta']],
					scroll: 2,
				}),
		);

		expect(output.text()).toContain('Viewing 3-4 of 5 results');
	});

	it('supports page, home, and end data table navigation keys', async () => {
		const output = createMemoryOutput();

		const pageDownResult = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.pageDown, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second'], ['Third'], ['Fourth']],
					scroll: 2,
				}),
		);

		const endHomeResult = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.end[0], Key.home[0], Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second'], ['Third'], ['Fourth']],
					scroll: 2,
				}),
		);

		expect(pageDownResult).toBe(2);
		expect(endHomeResult).toBe(0);
	});

	it('supports tab and reverse-tab data table navigation keys', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.tab, Key.tab, Key.shiftTab, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second'], ['Third']],
				}),
		);

		expect(result).toBe(1);
	});

	it('supports control-key data table navigation', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.ctrlN, Key.ctrlN, Key.ctrlP, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second'], ['Third']],
				}),
		);

		expect(result).toBe(1);
	});

	it('supports control-line data table navigation keys', async () => {
		const output = createMemoryOutput();

		const endResult = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.ctrlE, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second'], ['Third']],
				}),
		);

		const homeResult = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.ctrlE, Key.ctrlA, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second'], ['Third']],
				}),
		);

		expect(endResult).toBe(2);
		expect(homeResult).toBe(0);
	});

	it('wraps data table navigation from the first row to the last row', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.up, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second'], ['Third']],
				}),
		);

		expect(result).toBe(2);
	});

	it('normalizes fractional scroll sizes for data table page navigation', async () => {
		const output = createMemoryOutput();

		const result = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.pageDown, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				datatable({
					message: 'Pick row',
					headers: ['Name'],
					rows: [['First'], ['Second'], ['Third'], ['Fourth']],
					scroll: 2.9,
				}),
		);

		const latestFrame = output.text().split('Pick row\n').at(-2) ?? '';

		expect(result).toBe(2);
		expect(latestFrame).not.toContain('First');
		expect(latestFrame).not.toContain('Second');
		expect(latestFrame).toContain('Third');
		expect(latestFrame).toContain('Fourth');
	});

	it('works from form builders', async () => {
		const output = createMemoryOutput();

		const responses = await withPromptEnvironment(
			{
				input: createScriptedInput([Key.down, Key.enter]),
				output,
				error: output,
				interactive: true,
			},
			() =>
				form()
					.datatable(
						{
							message: 'Pick project',
							rows: [
								{ Name: 'Alpha', value: 'alpha' },
								{ Name: 'Beta', value: 'beta' },
							],
						},
						'project',
					)
					.submit(),
		);

		expect(responses.project).toBe('beta');
	});
});

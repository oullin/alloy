import { describe, expect, it } from 'vite-plus/test';
import { alert, createMemoryOutput, error, grid, info, intro, note, outro, parseAnsiText, table, warning, withPromptEnvironment } from '#console/index';
import { renderGrid } from '#console/output/grid';

describe('output helpers', () => {
	it('renders note-style helpers without returning a value', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			expect(note('Hello')).toBeUndefined();
			expect(error('Nope')).toBeUndefined();
			expect(warning('Careful')).toBeUndefined();
			expect(alert('Heads up')).toBeUndefined();
			expect(info('Facts')).toBeUndefined();
			expect(intro('Start')).toBeUndefined();
			expect(outro('Done')).toBeUndefined();
		});

		expect(output.text()).toContain('Hello');
		expect(output.text()).toContain('Done');
	});

	it('renders plain multiline notes without symbol prefixes', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			note('Hello\nWorld');
		});

		expect(output.text()).toBe(' Hello\n World\n');
	});

	it('renders typed note helpers with line-oriented styling', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			warning('Careful');
			error('Nope');
			alert('Heads up');
			info('Facts');
			intro('Start\nLonger');
			outro('Done');
		});

		expect(parseAnsiText(output.text())).toBe(' Careful\n Nope\n  Heads up \n Facts\n  Start  \n  Longer \n  Done \n');
		expect(output.text()).toContain('\u001B[33m Careful\u001B[39m');
		expect(output.text()).toContain('\u001B[31m Nope\u001B[39m');
		expect(output.text()).toContain('\u001B[41m\u001B[37m Heads up \u001B[39m\u001B[49m');
		expect(output.text()).toContain('\u001B[32m Facts\u001B[39m');
		expect(output.text()).toContain('\u001B[46m\u001B[30m Start  \u001B[39m\u001B[49m');
	});

	it('pads intro and outro lines by visible width', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			intro('東京\nA');
		});

		expect(parseAnsiText(output.text())).toBe('  東京 \n  A    \n');
	});

	it('renders table and boxed grid helpers without returning a value', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			expect(table(['Name'], [['Ollin']])).toBeUndefined();
			expect(grid(['A', 'B'], 80)).toBeUndefined();
		});

		expect(output.text()).toContain('| Name');
		expect(output.text()).toContain('┌───┬───┐');
		expect(output.text()).toContain('│ A │ B │');
	});

	it('keeps grid rendering finite for invalid widths', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			expect(grid(['A', 'B'], 0)).toBeUndefined();
			expect(grid(['C', 'D'], -2)).toBeUndefined();
		});

		expect(output.text()).toContain('│ A │');
		expect(output.text()).toContain('│ B │');
		expect(output.text()).toContain('│ C │');
		expect(output.text()).toContain('│ D │');
	});

	it('normalizes fractional grid widths without overlapping boxed rows', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			expect(grid(['A', 'B', 'C', 'D'], 2.5)).toBeUndefined();
		});

		expect(output.text()).toBe(' ┌───┐\n │ A │\n ├───┤\n │ B │\n ├───┤\n │ C │\n ├───┤\n │ D │\n └───┘\n');
	});

	it('renders balanced grid rows with table separators', () => {
		expect(renderGrid(['component-alpha', 'api-client', 'theme-box', 'status-task', 'prompt-input', 'stream-log'], 50)).toBe(
			[
				' ┌─────────────────┬─────────────┐',
				' │ component-alpha │ api-client  │',
				' ├─────────────────┼─────────────┤',
				' │ theme-box       │ status-task │',
				' ├─────────────────┼─────────────┤',
				' │ prompt-input    │ stream-log  │',
				' └─────────────────┴─────────────┘',
			].join('\n'),
		);
	});

	it('does not render empty grids', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			grid([]);
		});

		expect(output.text()).toBe('');
	});
});

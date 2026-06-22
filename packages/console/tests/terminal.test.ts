import { describe, expect, it } from 'vite-plus/test';

import {
	backgroundColor,
	clear,
	clearTerminal,
	createMemoryOutput,
	cursorToStart,
	eraseLine,
	erasePreviousLines,
	foregroundColor,
	hideCursor,
	setTerminalTitle,
	showCursor,
	supportsTrueColor,
	title,
	withPromptEnvironment,
} from '#console/index';

describe('terminal helpers', () => {
	it('writes clear terminal control sequences', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			clearTerminal();
			expect(clear()).toBeUndefined();
		});

		expect(output.text()).toBe('\u001B[H\u001B[J\u001B[H\u001B[J');
	});

	it('writes terminal title control sequences', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			setTerminalTitle('Demo');
			expect(title('Prompt')).toBeUndefined();
		});

		expect(output.text()).toBe('\u001B]0;Demo\u0007\u001B]0;Prompt\u0007');
	});

	it('writes cursor and erase-line control sequences', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			cursorToStart();
			eraseLine();
			erasePreviousLines(2);
			erasePreviousLines(-1);
			hideCursor();
			showCursor();
		});

		expect(output.text()).toBe('\r\u001B[2K\u001B[1A\u001B[2K\u001B[1A\u001B[2K\u001B[?25l\u001B[?25h');
	});

	it('detects true color terminal modes', () => {
		expect(supportsTrueColor('truecolor')).toBe(true);
		expect(supportsTrueColor('24bit')).toBe(true);
		expect(supportsTrueColor('256color')).toBe(false);
		expect(supportsTrueColor(undefined)).toBe(supportsTrueColor(process.env.COLORTERM));
	});

	it('parses terminal colors with default fallbacks', () => {
		expect(foregroundColor([1, 2, 3])).toEqual([1, 2, 3]);
		expect(backgroundColor([4, 5, 6])).toEqual([4, 5, 6]);
		expect(foregroundColor(['1', 2, 3])).toEqual([204, 204, 204]);
		expect(backgroundColor([0, 0, 300])).toEqual([0, 0, 0]);
	});
});

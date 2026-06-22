import { describe, expect, it } from 'vite-plus/test';
import { parseAnsiSegments, parseAnsiText, truncate, visibleWidth, wrap } from '#console/index';

describe('string utilities', () => {
	it('measures visible width without ANSI control sequences', () => {
		expect(visibleWidth('\u001B[31mPrompts\u001B[39m')).toBe(7);
	});

	it('truncates by visible width', () => {
		expect(truncate('Ollin Prompts', 10)).toBe('Ollin P...');
	});

	it('closes ANSI styles when truncating styled text', () => {
		const result = truncate('\u001B[31mOllin Prompts\u001B[39m', 10);

		expect(parseAnsiText(result)).toBe('Ollin P...');
		expect(result).toContain('\u001B[31m');
		expect(result).toContain('\u001B[0m...');
		expect(result.endsWith('\u001B[39m')).toBe(false);
	});

	it('clips truncation markers by visible width', () => {
		expect(truncate('Ollin Prompts', 1, '…')).toBe('…');
		expect(truncate('Ollin Prompts', 1, '東京')).toBe('');
	});

	it('wraps words to the requested width', () => {
		expect(wrap('Ollin Prompts', 7)).toEqual(['Ollin', 'Prompts']);
	});

	it('wraps wide characters without emitting empty lines', () => {
		expect(wrap('😀', 1)).toEqual(['😀']);
	});

	it('strips ANSI text', () => {
		expect(parseAnsiText('\u001B[32mDone\u001B[39m')).toBe('Done');
	});

	it('parses plain text into a single ANSI segment', () => {
		expect(parseAnsiSegments('Hello, World!')).toEqual([{ text: 'Hello, World!', codes: '' }]);
	});

	it('parses styled and unstyled ANSI segments', () => {
		expect(parseAnsiSegments('Hello \u001B[1mBold\u001B[0m World')).toEqual([
			{ text: 'Hello ', codes: '' },
			{ text: 'Bold', codes: '\u001B[1m' },
			{ text: ' World', codes: '' },
		]);
	});

	it('parses 24-bit ANSI color segments', () => {
		expect(parseAnsiSegments('\u001B[38;2;255;100;50mColored\u001B[0m')).toEqual([{ text: 'Colored', codes: '\u001B[38;2;255;100;50m' }]);
	});

	it('preserves ANSI codes across wrapped lines', () => {
		const result = wrap('\u001B[31mHello World\u001B[0m', 5);

		expect(result).toHaveLength(2);
		expect(result[0]).toContain('\u001B[31m');
		expect(result[0]).toContain('Hello');
		expect(result[0]?.endsWith('\u001B[0m')).toBe(true);
		expect(result[1]).toContain('\u001B[31m');
		expect(result[1]).toContain('World');
		expect(result[1]?.endsWith('\u001B[0m')).toBe(true);
	});

	it('wraps multiple ANSI color segments onto separate lines', () => {
		const result = wrap('\u001B[31mRed\u001B[0m \u001B[32mGreen\u001B[0m \u001B[34mBlue\u001B[0m', 5);

		expect(result).toHaveLength(3);
		expect(result[0]).toContain('Red');
		expect(result[1]).toContain('Green');
		expect(result[2]).toContain('Blue');
	});
});

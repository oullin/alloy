import { describe, expect, it } from 'vite-plus/test';

import { parseAnsiSegments, parseAnsiText, truncate, visibleWidth, wrap } from '#console/index';

describe('string utility edge cases', () => {
	it('measures combining marks, emoji, and wide CJK characters by terminal width', () => {
		expect(visibleWidth('e\u0301')).toBe(1);
		expect(visibleWidth('東京')).toBe(4);
		expect(visibleWidth('A😀界')).toBe(5);
		expect(visibleWidth('\u001B[31m東京\u001B[39m')).toBe(4);
	});

	it('truncates wide and combining text without splitting visible cells', () => {
		expect(truncate('東京abc', 5, '…')).toBe('東京…');
		expect(truncate('e\u0301clair', 4, '…')).toBe('e\u0301cl…');
		expect(truncate('東京', 1, '…')).toBe('…');
	});

	it('preserves active ANSI styles across truncation and strips them back to text', () => {
		const value = '\u001B[31mRed \u001B[1mbold text\u001B[22m plain\u001B[39m';
		const result = truncate(value, 12);

		expect(parseAnsiText(result)).toBe('Red bold ...');
		expect(result).toContain('\u001B[31m');
		expect(result).toContain('\u001B[1m');
		expect(result).toContain('\u001B[0m...');
	});

	it('parses reset style segments independently', () => {
		expect(parseAnsiSegments('\u001B[31mred\u001B[39m plain \u001B[1mbold\u001B[22m')).toEqual([
			{ text: 'red', codes: '\u001B[31m' },
			{ text: ' plain ', codes: '' },
			{ text: 'bold', codes: '\u001B[1m' },
		]);
	});

	it('wraps ANSI-colored wide words and closes styles on each emitted line', () => {
		const result = wrap('\u001B[32m東京大阪\u001B[0m plain', 4);

		expect(result.map(parseAnsiText)).toEqual(['東京', '大阪', 'plai', 'n']);
		expect(result[0]).toBe('\u001B[32m東京\u001B[0m');
		expect(result[1]).toBe('\u001B[32m大阪\u001B[0m');
	});
});

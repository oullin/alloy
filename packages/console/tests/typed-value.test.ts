import { describe, expect, it } from 'vite-plus/test';
import { applyTypedKey, Key } from '#console/index';
import { characterLength } from '#console/typed-value/characters';
import { visibleLineWindow } from '#console/typed-value/lines';

const apply = (keys: string[]): string => {
	let state = { cursor: 0, value: '' };

	for (const key of keys) {
		const next = applyTypedKey(state, key);

		state = { cursor: next.cursor, value: next.value };
	}

	return state.value;
};

const applyMultiline = (keys: string[]): { submitted: boolean; value: string } => {
	let state = { cursor: 0, value: '' };
	let submitted = false;

	for (const key of keys) {
		const next = applyTypedKey(state, key, true);

		state = { cursor: next.cursor, value: next.value };
		submitted = next.submitted;
	}

	return { submitted, value: state.value };
};

const applyMultilineState = (value: string, cursor: number, keys: string[], wrapWidth?: number): { cursor: number; value: string } => {
	let state = { cursor, value };

	for (const key of keys) {
		const next = applyTypedKey(state, key, true, wrapWidth);

		state = { cursor: next.cursor, value: next.value };
	}

	return state;
};

describe('typed value editing', () => {
	it('inserts buffered characters', () => {
		expect(apply(['Je', 'ss'])).toBe('Jess');
	});

	it('handles backspace and delete', () => {
		expect(apply(['J', 'e', 'z', Key.backspace, 's', 's'])).toBe('Jess');
		expect(apply(['J', 'e', 'z', Key.left, Key.delete, 's', 's'])).toBe('Jess');
	});

	it('keeps delete at the end of input unchanged', () => {
		expect(apply(['J', 'e', 's', 's', Key.delete])).toBe('Jess');
	});

	it('deletes typed input back to the line start', () => {
		expect(apply(['J', 'a', 'n', 'e', Key.left, Key.left, Key.ctrlU, 'D', 'o'])).toBe('Done');
	});

	it('supports emacs-style movement keys', () => {
		expect(apply(['J', 'z', 'e', Key.ctrlB, Key.ctrlH, Key.ctrlF, 's', 's'])).toBe('Jess');
	});

	it('marks typed values as cancelled without changing the current value', () => {
		expect(applyTypedKey({ cursor: 3, value: 'Ada' }, Key.ctrlC)).toEqual({
			cancelled: true,
			cursor: 3,
			submitted: false,
			value: 'Ada',
		});
	});

	it('clamps UTF-16 cursor offsets to character positions before editing', () => {
		expect(applyMultilineState('😀', '😀'.length, ['!'])).toEqual({
			cursor: 2,
			value: '😀!',
		});
	});

	it('deletes the previous word after non-BMP characters', () => {
		expect(applyMultilineState('😀 hello', characterLength('😀 hello'), [Key.optionBackspace])).toEqual({
			cursor: 2,
			value: '😀 ',
		});
	});

	it('deletes the previous punctuation group as a word boundary', () => {
		expect(applyMultilineState('hello --', characterLength('hello --'), [Key.optionBackspace])).toEqual({
			cursor: 6,
			value: 'hello ',
		});
	});

	it('moves to the start and end of a line', () => {
		expect(apply(['A', 'r', Key.home[0], 'J', Key.end[0], 'c', 'h', 'e', 'r'])).toBe('JArcher');
	});

	it('submits multiline input with ctrl-d', () => {
		expect(applyMultiline(['A', Key.enter, 'B', Key.ctrlD])).toEqual({
			submitted: true,
			value: 'A\nB',
		});
	});

	it('moves the cursor between textarea lines', () => {
		expect(applyMultilineState('abc\nde\nfghi', 6, [Key.up])).toEqual({
			cursor: 2,
			value: 'abc\nde\nfghi',
		});

		expect(applyMultilineState('abc\nde\nfghi', 2, [Key.down])).toEqual({
			cursor: 6,
			value: 'abc\nde\nfghi',
		});

		expect(applyMultilineState('abc\nde\nfghi', 6, [Key.down])).toEqual({
			cursor: 9,
			value: 'abc\nde\nfghi',
		});
	});

	it('moves between textarea lines with control navigation keys', () => {
		expect(applyMultilineState('abc\nde\nfghi', 6, [Key.ctrlP])).toEqual({
			cursor: 2,
			value: 'abc\nde\nfghi',
		});

		expect(applyMultilineState('abc\nde\nfghi', 2, [Key.ctrlN])).toEqual({
			cursor: 6,
			value: 'abc\nde\nfghi',
		});
	});

	it('moves the textarea cursor between wrapped rows', () => {
		expect(applyMultilineState('abcdefghijkl', 7, [Key.up], 5)).toEqual({
			cursor: 2,
			value: 'abcdefghijkl',
		});

		expect(applyMultilineState('abcdefghijkl', 2, [Key.down], 5)).toEqual({
			cursor: 7,
			value: 'abcdefghijkl',
		});

		expect(applyMultilineState('abcdefghijkl', 11, [Key.up], 5)).toEqual({
			cursor: 6,
			value: 'abcdefghijkl',
		});
	});

	it('moves to textarea line boundaries', () => {
		expect(applyMultilineState('abc\nde\nfghi', 5, [Key.home[0]])).toEqual({
			cursor: 4,
			value: 'abc\nde\nfghi',
		});

		expect(applyMultilineState('abc\nde\nfghi', 5, [Key.end[0]])).toEqual({
			cursor: 6,
			value: 'abc\nde\nfghi',
		});

		expect(applyMultilineState('abc\nde\nfghi', 9, [Key.ctrlA, 'X', Key.ctrlE, 'Y'])).toEqual({
			cursor: 13,
			value: 'abc\nde\nXfghiY',
		});
	});

	it('deletes textarea input back to the current line start', () => {
		expect(applyMultilineState('abc\nde\nfghi', 9, [Key.ctrlU])).toEqual({
			cursor: 7,
			value: 'abc\nde\nhi',
		});
	});

	it('reports visible textarea line windows', () => {
		expect(visibleLineWindow('A\nB\nC', 5, 2)).toEqual({
			lines: ['B', 'C'],
			start: 1,
			total: 3,
		});
	});

	it('reports wrapped textarea line windows', () => {
		expect(visibleLineWindow('abcdefghijkl', 12, 2, 5)).toEqual({
			lines: ['fghij', 'kl'],
			start: 1,
			total: 3,
		});
	});
});

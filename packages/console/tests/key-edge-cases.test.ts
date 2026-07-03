import { describe, expect, it } from 'vite-plus/test';

import { Key, keyFromEvent, oneOf } from '#console/index';
import { isCompleteRawKey, isPartialRawKey, normalizeRawKey } from '#console/environment/raw-key/normalize';

describe('keyboard edge cases', () => {
	it('maps arrow, escape, delete, and shifted arrow event names', () => {
		expect(keyFromEvent({ name: 'up' })).toBe(Key.up);
		expect(keyFromEvent({ name: 'down' })).toBe(Key.down);
		expect(keyFromEvent({ name: 'left' })).toBe(Key.left);
		expect(keyFromEvent({ name: 'right' })).toBe(Key.right);
		expect(keyFromEvent({ name: 'escape' })).toBe(Key.escape);
		expect(keyFromEvent({ name: 'delete' })).toBe(Key.delete);
		expect(keyFromEvent({ name: 'up', shift: true })).toBe(Key.shiftUp);
		expect(keyFromEvent({ name: 'down', shift: true })).toBe(Key.shiftDown);
	});

	it('maps common ctrl combinations case-insensitively and falls back to sequences', () => {
		expect(keyFromEvent({ ctrl: true, name: 'A' })).toBe(Key.ctrlA);
		expect(keyFromEvent({ ctrl: true, name: 'd' })).toBe(Key.ctrlD);
		expect(keyFromEvent({ ctrl: true, name: 'e' })).toBe(Key.ctrlE);
		expect(keyFromEvent({ sequence: '\r' })).toBe(Key.enter);
		expect(keyFromEvent({ sequence: 'x' })).toBe('x');
		expect(keyFromEvent({ name: 'unknown' })).toBe('unknown');
	});

	it('normalizes raw key aliases and detects partial escape sequences', () => {
		expect(normalizeRawKey('\r')).toBe(Key.enter);
		expect(normalizeRawKey('\u0008')).toBe(Key.backspace);
		expect(isPartialRawKey('\u001B[')).toBe(true);
		expect(isCompleteRawKey('\u001B[')).toBe(false);
		expect(isCompleteRawKey(Key.up)).toBe(true);
		expect(isCompleteRawKey('x')).toBe(true);
	});

	it('rejects invalid key matcher lists through the validator layer', () => {
		expect(oneOf([Key.up, Key.down, Key.home], Key.home[1])).toBe(Key.home[1]);
		expect(() => oneOf([Key.enter, 1 as unknown as string], Key.enter)).toThrow('Key match values must be strings or string arrays.');
	});
});

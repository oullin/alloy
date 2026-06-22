import { describe, expect, it } from 'vitest';
import { Key, keyFromEvent, oneOf } from '#console/index';

describe('keyboard mapping', () => {
	it('maps named keys', () => {
		expect(keyFromEvent({ name: 'enter' })).toBe(Key.enter);
	});

	it('maps terminal key name aliases', () => {
		expect(keyFromEvent({ name: 'return' })).toBe(Key.enter);
		expect(keyFromEvent({ name: 'pageup' })).toBe(Key.pageUp);
		expect(keyFromEvent({ name: 'pagedown' })).toBe(Key.pageDown);
	});

	it('maps space and newline sequences', () => {
		expect(keyFromEvent({ sequence: ' ' })).toBe(Key.space);
		expect(keyFromEvent({ sequence: '\n' })).toBe(Key.enter);
	});

	it('maps control-modified key names', () => {
		expect(keyFromEvent({ ctrl: true, name: 'c' })).toBe(Key.ctrlC);
		expect(keyFromEvent({ ctrl: true, name: 'n' })).toBe(Key.ctrlN);
		expect(keyFromEvent({ ctrl: true, name: 'U' })).toBe(Key.ctrlU);
	});

	it('maps option backspace events', () => {
		expect(keyFromEvent({ meta: true, name: 'backspace' })).toBe(Key.optionBackspace);
	});

	it('maps shifted tab events', () => {
		expect(keyFromEvent({ name: 'tab', shift: true })).toBe(Key.shiftTab);
	});

	it('matches nested key variants', () => {
		expect(oneOf([Key.home, Key.end], Key.home[2])).toBe(Key.home[2]);
		expect(oneOf([Key.home, Key.end], Key.end[3])).toBe(Key.end[3]);
		expect(oneOf([Key.home, Key.end], Key.enter)).toBeUndefined();
	});

	it('matches scalar and grouped keys through the validator layer', () => {
		expect(oneOf([Key.enter, Key.home], Key.enter)).toBe(Key.enter);
		expect(oneOf([Key.enter, Key.home], Key.home[0])).toBe(Key.home[0]);
	});
});

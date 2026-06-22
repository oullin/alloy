import { Key } from '#console/key';

const confirmToggleKeys = new Set([
	Key.tab,
	Key.up,
	Key.upArrow,
	Key.down,
	Key.downArrow,
	Key.left,
	Key.leftArrow,
	Key.right,
	Key.rightArrow,
	Key.ctrlP,
	Key.ctrlF,
	Key.ctrlN,
	Key.ctrlB,
	'h',
	'j',
	'k',
	'l',
]);

export const isConfirmToggleKey = (key: string): boolean => confirmToggleKeys.has(key);

export const confirmDirectValue = (key: string): boolean | null => {
	const normalizedKey = key.toLowerCase();

	if (normalizedKey === 'y') {
		return true;
	}

	if (normalizedKey === 'n') {
		return false;
	}

	return null;
};

import { Key } from '#console/key/constants';
import { ctrlKeyMap, namedKeyMap, shiftedKeyMap } from '#console/key/event/maps';
import { firstKeyValue, keyValueFromName } from '#console/key/event/value';
import { parseKeyboardEvent } from '#console/key/validators/event';
import type { KeyboardEventLike, KeyName } from '#console/key/types';

export const keyFromEvent = (event: KeyboardEventLike): KeyName | (string & {}) => {
	const parsedEvent = parseKeyboardEvent(event);

	if (parsedEvent.ctrl && parsedEvent.name) {
		const key = ctrlKeyMap[parsedEvent.name.toLowerCase()];

		if (key !== undefined) {
			return key;
		}
	}

	if (parsedEvent.meta && parsedEvent.name === 'backspace') {
		return Key.optionBackspace;
	}

	if (parsedEvent.shift && parsedEvent.name?.toLowerCase() === 'tab') {
		return Key.shiftTab;
	}

	if (parsedEvent.shift && parsedEvent.name) {
		const key = shiftedKeyMap[parsedEvent.name.toLowerCase()];

		if (key !== undefined) {
			return key;
		}
	}

	if (parsedEvent.name) {
		const key = namedKeyMap[parsedEvent.name.toLowerCase()];

		if (key !== undefined) {
			return key;
		}
	}

	if (parsedEvent.name) {
		const key = keyValueFromName(parsedEvent.name);

		if (key !== undefined) {
			return firstKeyValue(key);
		}
	}

	if (parsedEvent.sequence === ' ') {
		return Key.space;
	}

	if (parsedEvent.sequence === '\r' || parsedEvent.sequence === '\n') {
		return Key.enter;
	}

	return parsedEvent.name ?? parsedEvent.sequence ?? '';
};

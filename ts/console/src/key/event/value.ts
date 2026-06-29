import { Key } from '#console/key/constants';
import { parseFirstKeyValue, parseKeyLookupName } from '#console/key/validators/value';
import type { KeyValue } from '#console/key/types';

export const keyValueFromName = (name: string): KeyValue | undefined => {
	const keyName = parseKeyLookupName(name);

	if (keyName === undefined) {
		return undefined;
	}

	return Key[keyName];
};

export const firstKeyValue = (key: KeyValue): string => parseFirstKeyValue(key);

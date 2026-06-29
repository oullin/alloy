import { z } from 'zod';
import { Key } from '#console/key/constants';
import type { KeyValue } from '#console/key/types';

type KeyLookupName = keyof typeof Key;

const keyLookupNameSchema = z.enum(Object.keys(Key) as [KeyLookupName, ...KeyLookupName[]]);
const keyValueListSchema = z.array(z.string()).readonly();

export const isKeyValueList = (value: KeyValue): value is readonly string[] => {
	return keyValueListSchema.safeParse(value).success;
};

export const parseKnownKeyValues = (): string[] => {
	return Object.values(Key).flatMap((value) => (isKeyValueList(value) ? [...value] : [value]));
};

export const parseFirstKeyValue = (value: KeyValue): string => {
	return isKeyValueList(value) ? value[0] : value;
};

export const parseKeyLookupName = (value: unknown): KeyLookupName | undefined => {
	const parsed = keyLookupNameSchema.safeParse(value);

	return parsed.success ? parsed.data : undefined;
};

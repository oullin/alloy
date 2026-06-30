import { parseKeyMatchValues } from '#console/key/validators/match';
import type { KeyValue } from '#console/key/types';

export const oneOf = (keys: KeyValue[], match: string): string | undefined => {
	return parseKeyMatchValues(keys).includes(match) ? match : undefined;
};

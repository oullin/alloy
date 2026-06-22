import { z } from 'zod';
import { isStatusLabel } from '#console/form/builder/validators/status/common';
import { asyncIterableSchema, iterableSchema } from '#console/validators/iterable';

const streamSourceSchema = z.union([asyncIterableSchema<string>(), iterableSchema<string>()]);

export type ResolvedStreamFormArguments =
	| {
			kind: 'manual';
			name?: string;
	  }
	| {
			kind: 'source';
			name?: string;
			source: AsyncIterable<string> | Iterable<string>;
	  };

export const isStreamSource = (value: unknown): value is AsyncIterable<string> | Iterable<string> => {
	return streamSourceSchema.safeParse(value).success;
};

export const resolveStreamFormArguments = (sourceOrName?: AsyncIterable<string> | Iterable<string> | string, name?: string): ResolvedStreamFormArguments => {
	if (isStatusLabel(sourceOrName)) {
		return { kind: 'manual', name: sourceOrName };
	}

	if (!isStreamSource(sourceOrName)) {
		return { kind: 'manual' };
	}

	return { kind: 'source', name, source: sourceOrName };
};

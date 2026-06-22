import { z } from 'zod';
import { functionSchema } from '#console/validators/function';

export type WritableProcessStream = {
	write: NodeJS.WriteStream['write'];
};

const processOutputWriteSchema = functionSchema<NodeJS.WriteStream['write']>();

const writableProcessStreamSchema = z
	.object({
		write: processOutputWriteSchema,
	})
	.passthrough() as unknown as z.ZodType<WritableProcessStream>;

export const parseProcessOutputWrite = (value: unknown): NodeJS.WriteStream['write'] => {
	const parsed = processOutputWriteSchema.safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Process output writers must be functions.');
	}

	return value as NodeJS.WriteStream['write'];
};

export const parseWritableProcessStream = (value: unknown): WritableProcessStream => {
	const parsed = writableProcessStreamSchema.safeParse(value);

	if (!parsed.success) {
		throw new TypeError('Process output streams must include a write function.');
	}

	return value as WritableProcessStream;
};

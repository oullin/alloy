import { z } from 'zod';
import { functionSchema } from '#console/validators/function';

export type ProcessOutputCallback = (error?: Error | null) => void;

const processOutputEncodingSchema = z.string().refine((value): value is BufferEncoding => Buffer.isEncoding(value));

const processOutputCallbackSchema = functionSchema<ProcessOutputCallback>();

export const parseProcessOutputEncoding = (value: unknown): BufferEncoding | undefined => {
	const parsed = processOutputEncodingSchema.safeParse(value);

	return parsed.success ? parsed.data : undefined;
};

export const parseProcessOutputCallback = (encodingOrCallback: unknown, callback?: ProcessOutputCallback): ProcessOutputCallback | undefined => {
	const parsedEncodingCallback = processOutputCallbackSchema.safeParse(encodingOrCallback);

	if (parsedEncodingCallback.success) {
		return parsedEncodingCallback.data;
	}

	const parsedCallback = processOutputCallbackSchema.safeParse(callback);

	return parsedCallback.success ? parsedCallback.data : undefined;
};

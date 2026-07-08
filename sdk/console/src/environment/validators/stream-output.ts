import { z } from 'zod';
import { functionSchema } from '#console/validators/function';

const writableStreamSchema = z
	.object({
		write: functionSchema<NodeJS.WritableStream['write']>(),
	})
	.passthrough() as unknown as z.ZodType<NodeJS.WritableStream>;

export const parseWritableOutputStream = (stream: unknown): NodeJS.WritableStream => {
	const parsed = writableStreamSchema.safeParse(stream);

	if (!parsed.success) {
		throw new Error('Prompt output streams must include a write function.');
	}

	return stream as NodeJS.WritableStream;
};

import { z } from 'zod';
import { functionSchema } from '#console/validators/function';

export type StatusSignalTarget = {
	off(signal: string, listener: () => void): unknown;
	on(signal: string, listener: () => void): unknown;
};

const statusSignalTargetSchema = z
	.object({
		off: functionSchema<StatusSignalTarget['off']>(),
		on: functionSchema<StatusSignalTarget['on']>(),
	})
	.passthrough() as unknown as z.ZodType<StatusSignalTarget>;

export const parseStatusSignalTarget = (target: unknown): StatusSignalTarget => {
	const parsed = statusSignalTargetSchema.safeParse(target);

	if (!parsed.success) {
		throw new TypeError('Status signal targets must include on and off functions.');
	}

	return target as StatusSignalTarget;
};

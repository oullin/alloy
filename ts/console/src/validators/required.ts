import { z } from 'zod';

const requiredOptionSchema = z.union([z.boolean(), z.string()]).optional();
const requiredMessageSchema = z.string().min(1);
const invalidRequiredValueSchema = z.union([z.literal(''), z.literal(false), z.null(), z.undefined(), z.array(z.unknown()).length(0)]);

const isInvalidRequiredValue = (value: unknown): boolean => {
	return invalidRequiredValueSchema.safeParse(value).success;
};

export const requiredMessage = (value: unknown, required?: boolean | string): string | undefined => {
	const requiredOption = requiredOptionSchema.safeParse(required);

	if (!requiredOption.success) {
		return undefined;
	}

	const parsed = requiredOption.data;

	if (parsed === false || parsed === undefined || !isInvalidRequiredValue(value)) {
		return undefined;
	}

	const message = requiredMessageSchema.safeParse(parsed);

	return message.success ? message.data : 'Required.';
};

import { z } from 'zod';
import type { NotificationCommand } from '#console/output/notify/commands';

const notificationCommandBinSchema = z
	.string()
	.min(1)
	.regex(/^(?!-)[A-Za-z0-9_./-]+$/u);

const notificationCommandSchema = z.object({
	args: z.array(z.string()).default([]),
	bin: notificationCommandBinSchema,
});

const notificationCommandListSchema = z.array(notificationCommandSchema);

export const parseNotificationCommand = (command: unknown): NotificationCommand | undefined => {
	const parsed = notificationCommandSchema.safeParse(command);

	return parsed.success ? parsed.data : undefined;
};

export const parseNotificationCommands = (commands: unknown): NotificationCommand[] => {
	const parsed = notificationCommandListSchema.safeParse(commands);

	return parsed.success ? parsed.data : [];
};

export const parseNotificationCommandBin = (bin: unknown): string | undefined => {
	const parsed = notificationCommandBinSchema.safeParse(bin);

	return parsed.success ? parsed.data : undefined;
};

import { z } from 'zod';

const notificationOptionsSchema = z.object({
	body: z.string().default(''),
	icon: z.string().default(''),
	sound: z.string().default(''),
	subtitle: z.string().default(''),
	title: z.string(),
});

export type NotificationOptions = z.input<typeof notificationOptionsSchema>;

export type ResolvedNotificationOptions = z.output<typeof notificationOptionsSchema>;

export const parseNotificationOptions = (options: unknown): ResolvedNotificationOptions => {
	const parsed = notificationOptionsSchema.safeParse(options);

	if (!parsed.success) {
		throw new Error('Notification options must include a string title.');
	}

	return parsed.data;
};

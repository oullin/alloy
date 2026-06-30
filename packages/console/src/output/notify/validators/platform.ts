import { z } from 'zod';

const notificationPlatforms = ['aix', 'android', 'cygwin', 'darwin', 'freebsd', 'haiku', 'linux', 'netbsd', 'openbsd', 'sunos', 'win32'] as const;
const notificationPlatformSchema = z.enum(notificationPlatforms);

export type NotificationPlatform = z.infer<typeof notificationPlatformSchema>;

export const parseNotificationPlatform = (value: unknown): NotificationPlatform => {
	const parsed = notificationPlatformSchema.safeParse(value);

	return parsed.success ? parsed.data : 'freebsd';
};

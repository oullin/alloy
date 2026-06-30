import { platform } from 'node:process';
import { note } from '#console/output/notes';
import { notificationCommands } from '#console/output/notify/commands';
import { availableNotificationCommands, commandExists, executeNotificationCommand } from '#console/output/notify/executor';
import { parseNotificationPlatform } from '#console/output/notify/validators/platform';
import type { NotificationCommand, NotificationPlatform } from '#console/output/notify/commands';
import type { NotificationRuntime } from '#console/output/notify/executor';

export type { NotificationRuntime };
export type { NotificationCommand };
export type { NotificationPlatform };
export type { NotificationOptions } from '#console/output/notify/validators/options';
export { commandExists, executeNotificationCommand, notificationCommands };

export const notificationCommand = (targetPlatform: NotificationPlatform, title: string, body = '', subtitle = '', sound = '', icon = ''): NotificationCommand | null => {
	return notificationCommands(targetPlatform, { body, icon, sound, subtitle, title }).at(0) ?? null;
};

export const notifyForPlatform = (targetPlatform: NotificationPlatform, title: string, body = '', subtitle = '', sound = '', icon = '', runtime: NotificationRuntime = {}): boolean => {
	const commands = notificationCommands(targetPlatform, { body, icon, sound, subtitle, title });
	const availableCommands = availableNotificationCommands(targetPlatform, commands, runtime.commandExists ?? commandExists);
	const execute = runtime.execute ?? executeNotificationCommand;

	for (const command of availableCommands) {
		if (execute(command)) {
			return true;
		}
	}

	note(body ? `${title}: ${body}` : title, 'info');

	return false;
};

export const notify = (title: string, body = '', subtitle = '', sound = '', icon = ''): void => {
	notifyForPlatform(parseNotificationPlatform(platform), title, body, subtitle, sound, icon);
};

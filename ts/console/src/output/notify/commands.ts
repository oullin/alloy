import { parseNotificationOptions } from '#console/output/notify/validators/options';
import type { NotificationPlatform } from '#console/output/notify/validators/platform';
import type { ResolvedNotificationOptions } from '#console/output/notify/validators/options';

export type NotificationCommand = {
	args: string[];
	bin: string;
};

export type { NotificationPlatform } from '#console/output/notify/validators/platform';

const macOsNotificationCommand = (options: ResolvedNotificationOptions): NotificationCommand => {
	const script = [
		'display notification',
		JSON.stringify(options.body),
		'with title',
		JSON.stringify(options.title),
		options.subtitle ? `subtitle ${JSON.stringify(options.subtitle)}` : '',
		options.sound ? `sound name ${JSON.stringify(options.sound)}` : '',
	]
		.filter(Boolean)
		.join(' ');

	return { args: ['-e', script], bin: 'osascript' };
};

const notifySendCommand = (options: ResolvedNotificationOptions): NotificationCommand => {
	const args = options.icon ? ['--icon', options.icon, options.title] : [options.title];

	if (options.body) {
		args.push(options.body);
	}

	return { args, bin: 'notify-send' };
};

const kdialogCommand = (options: ResolvedNotificationOptions): NotificationCommand => {
	const message = options.body ? `${options.title}: ${options.body}` : options.title;
	const args = ['--passivepopup', message, '5', '--title', options.title];

	return { args, bin: 'kdialog' };
};

export const notificationCommands = (targetPlatform: NotificationPlatform, options: unknown): NotificationCommand[] => {
	const parsed = parseNotificationOptions(options);

	if (targetPlatform === 'darwin') {
		return [macOsNotificationCommand(parsed)];
	}

	if (targetPlatform === 'linux') {
		return [notifySendCommand(parsed), kdialogCommand(parsed)];
	}

	return [];
};

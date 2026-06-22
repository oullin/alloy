import { describe, expect, expectTypeOf, it } from 'vite-plus/test';
import { createMemoryOutput, notificationCommand, notificationCommands, notify, notifyForPlatform, withPromptEnvironment } from '#console/index';
import type { NotificationCommand } from '#console/index';

describe('notify helper', () => {
	it('builds macOS notification commands', () => {
		const command = notificationCommand('darwin', 'Deploy', 'Done', 'Preview', 'Glass');

		expect(command?.bin).toBe('osascript');
		expect(command?.args[0]).toBe('-e');
		expect(command?.args[1]).toContain('display notification "Done"');
		expect(command?.args[1]).toContain('with title "Deploy"');
		expect(command?.args[1]).toContain('subtitle "Preview"');
		expect(command?.args[1]).toContain('sound name "Glass"');
	});

	it('builds Linux notification commands', () => {
		expect(notificationCommand('linux', 'Deploy', 'Done')).toEqual({
			args: ['Deploy', 'Done'],
			bin: 'notify-send',
		});

		expect(notificationCommand('linux', 'Deploy')).toEqual({
			args: ['Deploy'],
			bin: 'notify-send',
		});

		expect(notificationCommand('linux', 'Deploy', 'Done', '', '', '/tmp/icon.png')).toEqual({
			args: ['--icon', '/tmp/icon.png', 'Deploy', 'Done'],
			bin: 'notify-send',
		});
	});

	it('builds Linux fallback notification commands in availability order', () => {
		expect(notificationCommands('linux', { title: 'Deploy', body: 'Done' })).toEqual([
			{
				args: ['Deploy', 'Done'],
				bin: 'notify-send',
			},
			{
				args: ['--passivepopup', 'Deploy: Done', '5', '--title', 'Deploy'],
				bin: 'kdialog',
			},
		]);
	});

	it('falls back to note output on unsupported platforms', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			expect(notifyForPlatform('freebsd', 'Deploy', 'Done')).toBe(false);
		});

		expect(output.text()).toContain('Deploy: Done');
	});

	it('does not expose a public notification return value', () => {
		expectTypeOf(notify).returns.toBeVoid();
	});

	it('executes the first available Linux notifier and reports the process result', () => {
		const executed: NotificationCommand[] = [];

		const result = notifyForPlatform('linux', 'Deploy', 'Done', '', '', '', {
			commandExists: (bin) => bin === 'kdialog',
			execute: (command) => {
				executed.push(command);

				return true;
			},
		});

		expect(result).toBe(true);
		expect(executed).toEqual([
			{
				args: ['--passivepopup', 'Deploy: Done', '5', '--title', 'Deploy'],
				bin: 'kdialog',
			},
		]);
	});

	it('returns false when notification command execution fails', () => {
		const result = notifyForPlatform('darwin', 'Deploy', 'Done', '', '', '', {
			execute: () => false,
		});

		expect(result).toBe(false);
	});
});

import { describe, expect, it } from 'vite-plus/test';
import { EventEmitter } from 'node:events';
import { configurePrompts, createMemoryOutput, createScriptedInput, promptEnvironment, withPromptEnvironment } from '#console/environment';
import { readRawKey } from '#console/environment/raw-key';
import { Key } from '#console/key';

class FakeRawInput extends EventEmitter {
	isRaw = false;
	isTTY = true;
	paused = 0;
	rawModes: boolean[] = [];
	resumed = 0;

	pause(): void {
		this.paused += 1;
	}

	resume(): void {
		this.resumed += 1;
	}

	setRawMode(mode: boolean): void {
		this.rawModes.push(mode);
		this.isRaw = mode;
	}
}

describe('prompt environment', () => {
	it('restores the previous environment when a scoped callback fails', async () => {
		const output = createMemoryOutput();
		const error = createMemoryOutput();
		const previous = promptEnvironment();

		configurePrompts({ output, error, interactive: false });

		const scopedOutput = createMemoryOutput();

		await expect(
			withPromptEnvironment({ output: scopedOutput, interactive: true }, async () => {
				expect(promptEnvironment().output).toBe(scopedOutput);
				expect(promptEnvironment().error).toBe(error);
				expect(promptEnvironment().interactive).toBe(true);

				throw new Error('failed');
			}),
		).rejects.toThrow('failed');

		expect(promptEnvironment().output).toBe(output);
		expect(promptEnvironment().error).toBe(error);
		expect(promptEnvironment().interactive).toBe(false);

		configurePrompts(previous);
	});

	it('merges partial prompt configuration without replacing unspecified concerns', () => {
		const previous = promptEnvironment();
		const output = createMemoryOutput();
		const error = createMemoryOutput();

		configurePrompts({ output, error, interactive: false });

		const nextOutput = createMemoryOutput();

		configurePrompts({ output: nextOutput });

		expect(promptEnvironment().output).toBe(nextOutput);
		expect(promptEnvironment().error).toBe(error);
		expect(promptEnvironment().interactive).toBe(false);

		configurePrompts(previous);
	});

	it('captures and clears memory output', () => {
		const output = createMemoryOutput();

		output.write('first');
		output.write(' second');

		expect(output.text()).toBe('first second');

		output.clear();

		expect(output.text()).toBe('');
	});

	it('drains scripted input independently for key and line reads', async () => {
		const input = createScriptedInput(['a', 'line']);

		await expect(input.readKey?.()).resolves.toBe('a');

		await expect(input.readLine?.('Question')).resolves.toBe('line');

		await expect(input.readKey?.()).resolves.toBeNull();

		await expect(input.readLine?.('Question')).resolves.toBe('');
	});

	it('restores raw input mode after reading a key', async () => {
		const input = new FakeRawInput();
		const key = readRawKey(input);

		expect(input.rawModes).toEqual([true]);
		expect(input.resumed).toBe(1);

		input.emit('data', Buffer.from('x'));

		await expect(key).resolves.toBe('x');

		expect(input.rawModes).toEqual([true, false]);
		expect(input.paused).toBe(1);
		expect(input.listenerCount('data')).toBe(0);
		expect(input.listenerCount('end')).toBe(0);
		expect(input.listenerCount('error')).toBe(0);
	});

	it('normalizes raw terminal key aliases', async () => {
		const input = new FakeRawInput();
		const enter = readRawKey(input);

		input.emit('data', Buffer.from('\r'));

		await expect(enter).resolves.toBe(Key.enter);

		const backspaceInput = new FakeRawInput();
		const backspace = readRawKey(backspaceInput);

		backspaceInput.emit('data', Buffer.from('\u0008'));

		await expect(backspace).resolves.toBe(Key.backspace);
	});

	it('buffers split raw terminal escape sequences', async () => {
		const input = new FakeRawInput();
		const key = readRawKey(input);

		input.emit('data', Buffer.from('\u001B'));
		input.emit('data', Buffer.from('['));
		input.emit('data', Buffer.from('A'));

		await expect(key).resolves.toBe(Key.up);

		expect(input.rawModes).toEqual([true, false]);
	});

	it('returns buffered partial escape sequences when input ends', async () => {
		const input = new FakeRawInput();
		const key = readRawKey(input);

		input.emit('data', Buffer.from('\u001B'));
		input.emit('end');

		await expect(key).resolves.toBe(Key.escape);
	});

	it('restores raw input mode when key input ends', async () => {
		const input = new FakeRawInput();
		const key = readRawKey(input);

		input.emit('end');

		await expect(key).resolves.toBeNull();

		expect(input.rawModes).toEqual([true, false]);
		expect(input.paused).toBe(1);
	});

	it('restores raw input mode when key input errors', async () => {
		const input = new FakeRawInput();
		const key = readRawKey(input);

		input.emit('error', new Error('read failed'));

		await expect(key).rejects.toThrow('read failed');

		expect(input.rawModes).toEqual([true, false]);
		expect(input.paused).toBe(1);
	});
});

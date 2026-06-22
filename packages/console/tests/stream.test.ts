import { describe, expect, it } from 'vitest';
import { createMemoryOutput, parseAnsiText, stream, withPromptEnvironment } from '#console/index';
import { streamFadeStyles } from '#console/status/stream/fade';
import { renderStreamFrame } from '#console/status/stream/render';

describe('stream helper', () => {
	it('appends streamed content and returns the accumulated value', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			const outputStream = stream();

			outputStream.append('hello');
			outputStream.write(' world');
			outputStream.close();

			expect(outputStream.value()).toBe('hello world');
			expect(outputStream.lines()).toEqual(['hello world']);
			expect(outputStream.closed()).toBe(true);
		});

		expect(output.text()).toContain('\u001B[?25l');
		expect(output.text()).toContain('\u001B[?25h');
		expect(parseAnsiText(output.text())).toContain(' hello world');
	});

	it('pipes iterable content through a stateful stream', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			await stream(['one', '\n', 'two']);
		});

		expect(parseAnsiText(output.text())).toContain(' one');
		expect(parseAnsiText(output.text())).toContain(' two');
	});

	it('closes piped streams after the source is exhausted', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			const outputStream = stream();

			await outputStream.pipe(['done']);

			expect(outputStream.closed()).toBe(true);
			expect(outputStream.value()).toBe('done');
			expect(() => outputStream.write(' again')).toThrow('Stream is closed.');
		});

		expect(output.text()).toContain('\u001B[?25h');
	});

	it('rejects writes after close without losing buffered content', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			const outputStream = stream();

			outputStream.write('finished');
			outputStream.close();

			expect(() => outputStream.append(' later')).toThrow('Stream is closed.');
			expect(outputStream.value()).toBe('finished');
		});

		expect(parseAnsiText(output.text())).toContain(' finished');
	});

	it('cannot be prompted for input', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			const outputStream = stream();

			expect(() => outputStream.prompt()).toThrow('Stream cannot be prompted');
		});
	});

	it('wraps rendered stream lines with a leading gutter', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			const outputStream = stream();

			outputStream.write('a '.repeat(40).trimEnd());

			expect(outputStream.lines().length).toBeGreaterThan(1);
		});

		expect(output.text()).toContain('\n ');
	});

	it('flushes pending stream chunks before closing', async () => {
		const output = createMemoryOutput();

		await withPromptEnvironment({ output, error: output }, async () => {
			const outputStream = stream();

			for (let index = 0; index < 12; index += 1) {
				outputStream.append(String(index));
			}

			outputStream.close();

			expect(outputStream.value()).toBe('01234567891011');
			expect(outputStream.closed()).toBe(true);
		});

		expect(parseAnsiText(output.text())).toContain(' 01234567891011');
	});

	it('renders pending stream chunks with fallback fade styles', () => {
		const frame = renderStreamFrame({
			fading: ['fresh', ' fading'],
			fadeStyles: streamFadeStyles({ trueColor: false }),
			value: 'stable ',
		});

		expect(parseAnsiText(frame)).toBe(' stable fresh fading\n');
		expect(frame).toContain('stable fresh\u001B[2m fading\u001B[22m');
	});

	it('renders true-color stream fade styles', () => {
		const frame = renderStreamFrame({
			fading: ['fresh', ' fading'],
			fadeStyles: streamFadeStyles({
				background: [0, 0, 0],
				foreground: [100, 50, 0],
				steps: 2,
				trueColor: true,
			}),
			value: '',
		});

		expect(frame).toContain('\u001B[38;2;100;50;0mfresh\u001B[0m');
		expect(frame).toContain('\u001B[38;2;50;25;0m fading\u001B[0m');
	});
});

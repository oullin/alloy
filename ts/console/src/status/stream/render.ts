import { wrap } from '#console/strings';
import type { StreamFadeStyle } from '#console/status/stream/fade';
import { parseStreamChunkLines } from '#console/status/stream/validators/chunk';
import { parseStreamFrameWidth } from '#console/status/stream/validators/fade';

type StreamFrameOptions = {
	fading?: string[];
	fadeStyles?: StreamFadeStyle[];
	value: string;
	width?: number;
};

const fadedValue = (value: string, fading: string[], styles: StreamFadeStyle[]): string => {
	return `${value}${fading.map((chunk, index) => (styles[index] ?? styles.at(-1) ?? ((text: string) => text))(chunk)).join('')}`;
};

export const streamLines = ({ value, width }: StreamFrameOptions): string[] => {
	const frameWidth = parseStreamFrameWidth(width, 60);

	return parseStreamChunkLines(value).flatMap((line) => wrap(line, frameWidth));
};

export const renderStreamFrame = (options: StreamFrameOptions): string => {
	const value = fadedValue(options.value, options.fading ?? [], options.fadeStyles ?? []);

	return `${streamLines({ ...options, value })
		.map((line) => ` ${line}`)
		.join('\n')}\n`;
};

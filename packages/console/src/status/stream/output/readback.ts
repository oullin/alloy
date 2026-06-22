import { streamLines } from '#console/status/stream/render';
import type { StreamBuffer } from '#console/status/stream/buffer';

export const streamBufferValue = (buffer: StreamBuffer): string => {
	return buffer.value();
};

export const streamBufferLines = (buffer: StreamBuffer): string[] => {
	return streamLines({ value: streamBufferValue(buffer) });
};

import { resolveProcessOutputWrite } from '#console/status/task/process-output/validators/writer';
import type { ProcessOutputBuffer } from '#console/status/task/process-output/buffer';

export const processOutputWriter = (buffer: ProcessOutputBuffer) => {
	return (chunk: string | Uint8Array, encodingOrCallback?: BufferEncoding | ((error?: Error | null) => void), callback?: (error?: Error | null) => void): boolean => {
		const resolved = resolveProcessOutputWrite(chunk, encodingOrCallback, callback);

		buffer.write(resolved.content);

		if (resolved.callback) {
			queueMicrotask(() => {
				resolved.callback?.();
			});
		}

		return true;
	};
};

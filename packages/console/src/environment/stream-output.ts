import { parsePromptOutputContent } from '#console/environment/validators/output';
import { parseWritableOutputStream } from '#console/environment/validators/stream-output';
import type { PromptOutput } from '#console/types';

export const outputFromStream = (stream: NodeJS.WritableStream): PromptOutput => {
	const writable = parseWritableOutputStream(stream);
	const write = writable.write.bind(writable);

	return {
		write(content: string): void {
			write(parsePromptOutputContent(content));
		},
	};
};

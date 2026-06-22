import { parsePromptOutputContent } from '#console/environment/validators/output';
import type { PromptOutput } from '#console/types';

export type MemoryOutput = PromptOutput & {
	clear(): void;
	text(): string;
};

export const createMemoryOutput = (): MemoryOutput => {
	let buffer = '';

	return {
		clear(): void {
			buffer = '';
		},
		write(content: string): void {
			buffer += parsePromptOutputContent(content);
		},
		text(): string {
			return buffer;
		},
	};
};

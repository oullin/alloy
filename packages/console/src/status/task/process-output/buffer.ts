import type { Logger } from '#console/status/task/logger';
import { parseProcessOutputChunkLines } from '#console/status/task/process-output/validators/chunk';

export class ProcessOutputBuffer {
	#pending = '';
	readonly #logger: Logger;

	constructor(logger: Logger) {
		this.#logger = logger;
	}

	write(chunk: string): void {
		const lines = parseProcessOutputChunkLines(`${this.#pending}${chunk}`);

		this.#pending = lines.pop() ?? '';

		for (const line of lines) {
			if (line.length > 0) {
				this.#logger.line(line);
			}
		}
	}

	flush(): void {
		if (this.#pending.length === 0) {
			return;
		}

		this.#logger.line(this.#pending);
		this.#pending = '';
	}
}

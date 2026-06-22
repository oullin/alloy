import { StatusSignalCleanup } from '#console/status/signals';
import { hideCursor, showCursor } from '#console/terminal';

export class StreamLifecycle {
	#closed = false;
	readonly #flush: () => void;
	readonly #signalCleanup: StatusSignalCleanup;

	constructor(flush: () => void) {
		this.#flush = flush;
		this.#signalCleanup = new StatusSignalCleanup(() => {
			this.close();
		});

		try {
			hideCursor();
			this.#signalCleanup.attach();
		} catch (error) {
			showCursor();

			throw error;
		}
	}

	close(): void {
		if (this.#closed) {
			return;
		}

		try {
			this.#flush();
		} finally {
			this.#closed = true;
			this.#signalCleanup.detach();
			showCursor();
		}
	}

	closed(): boolean {
		return this.#closed;
	}
}

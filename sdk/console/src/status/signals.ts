import { parseStatusSignalTarget } from '#console/status/validators/signal-target';
import type { StatusSignalTarget } from '#console/status/validators/signal-target';

const statusSignals = ['SIGINT', 'SIGTERM'] as const;

export type { StatusSignalTarget };

export class StatusSignalCleanup {
	#attached = false;
	#restored = false;
	readonly #restore: () => void;
	readonly #target: StatusSignalTarget;
	readonly #handleSignal = (): void => {
		this.restore();
	};

	constructor(restore: () => void, target: StatusSignalTarget = process) {
		this.#restore = restore;
		this.#target = parseStatusSignalTarget(target);
	}

	attach(): this {
		if (this.#attached) {
			return this;
		}

		for (const signal of statusSignals) {
			this.#target.on(signal, this.#handleSignal);
		}

		this.#attached = true;

		return this;
	}

	restore(): void {
		if (this.#restored) {
			return;
		}

		this.#restored = true;
		this.detach();
		this.#restore();
	}

	detach(): void {
		if (!this.#attached) {
			return;
		}

		for (const signal of statusSignals) {
			this.#target.off(signal, this.#handleSignal);
		}

		this.#attached = false;
	}
}

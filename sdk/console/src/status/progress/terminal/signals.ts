import { parseStatusSignalTarget } from '#console/status/validators/signal-target';
import type { StatusSignalTarget } from '#console/status/validators/signal-target';

const progressSignals = ['SIGINT', 'SIGTERM'] as const;

export type ProgressSignalTarget = StatusSignalTarget;

export class ProgressSignalHandlers {
	#attached = false;
	readonly #handleSignal: () => void;
	readonly #target: ProgressSignalTarget;

	constructor(target: ProgressSignalTarget, handleSignal: () => void) {
		this.#target = parseStatusSignalTarget(target);
		this.#handleSignal = handleSignal;
	}

	attach(): void {
		if (this.#attached) {
			return;
		}

		for (const signal of progressSignals) {
			this.#target.on(signal, this.#handleSignal);
		}

		this.#attached = true;
	}

	detach(): void {
		if (!this.#attached) {
			return;
		}

		for (const signal of progressSignals) {
			this.#target.off(signal, this.#handleSignal);
		}

		this.#attached = false;
	}
}

import { progressStateSnapshot } from '#console/status/progress/state/snapshot';
import { clampedProgressCurrent, nextProgressCurrent, progressRatio } from '#console/status/progress/state/value';
import { parseProgressTotal } from '#console/status/validators/progress';
import type { ProgressFrameState } from '#console/status/progress/render';
import type { ProgressStateSnapshot } from '#console/status/progress/state/snapshot';

export type { ProgressStateSnapshot } from '#console/status/progress/state/snapshot';

export class ProgressState {
	#current = 0;
	#hint: string;
	#label: string;
	#state: ProgressFrameState = 'active';
	readonly total: number;

	constructor(total: number, label: string, hint: string) {
		this.total = parseProgressTotal(total);
		this.#label = label;
		this.#hint = hint;
	}

	activate(): void {
		this.#state = 'active';
	}

	advance(step: number): void {
		this.#state = 'active';
		this.#current = nextProgressCurrent(this.#current, this.total, step);
	}

	fail(): void {
		this.#state = 'error';
	}

	finish(): void {
		this.#current = this.total;
		this.#state = 'submit';
	}

	hint(value: string): void {
		this.#hint = value;
	}

	label(value: string): void {
		this.#label = value;
	}

	current(): number {
		return clampedProgressCurrent(this.#current, this.total);
	}

	percentage(): number {
		return progressRatio(this.#current, this.total);
	}

	snapshot(): ProgressStateSnapshot {
		return progressStateSnapshot(this.#current, this.total, this.#label, this.#hint, this.#state);
	}
}

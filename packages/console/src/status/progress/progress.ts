import { ProgressController } from '#console/status/progress/controller';
import { setProgressHint, setProgressLabel } from '#console/status/progress/labels';
import { progressPromptError } from '#console/status/progress/prompt';
import { progressCurrent, progressPercentageValue, progressValue } from '#console/status/progress/readback';
import type { ProgressSignalTarget } from '#console/status/progress/terminal';

export type { ProgressSignalTarget } from '#console/status/progress/terminal';

/**
 * Terminal progress bar with a label, hint, and signal-aware cleanup.
 * Advance it manually or drive it through the `progress()` helper.
 */
export class Progress {
	readonly #controller: ProgressController;
	readonly total: number;
	#handleSignal = (): void => {
		this.fail();
	};

	constructor(total: number, message = 'Progress', hint = '', signalTarget: ProgressSignalTarget = process) {
		this.#controller = new ProgressController(total, message, hint, signalTarget, this.#handleSignal);
		this.total = this.#controller.state.total;
	}

	/** Renders the bar and begins listening for termination signals. */
	start(): void {
		this.#controller.activate();
	}

	/** Advances the bar by the given number of steps. */
	advance(step = 1): void {
		this.#controller.advance(step);
	}

	/** Completes the bar and restores the terminal. */
	finish(): void {
		this.#controller.finish();
	}

	/** Marks the bar as failed and restores the terminal. */
	fail(): void {
		this.#controller.fail();
	}

	/** Updates the bar's label. */
	label(value: string): this {
		setProgressLabel(this.#controller, value);

		return this;
	}

	/** Updates the bar's hint text. */
	hint(value: string): this {
		setProgressHint(this.#controller, value);

		return this;
	}

	/** Returns completion as a whole-number percentage. */
	percentage(): number {
		return progressPercentageValue(this.#controller);
	}

	/** Returns the number of completed steps. */
	current(): number {
		return progressCurrent(this.#controller);
	}

	value(): boolean {
		return progressValue();
	}

	prompt(): never {
		throw progressPromptError();
	}

	render(): void {
		this.#controller.render();
	}
}

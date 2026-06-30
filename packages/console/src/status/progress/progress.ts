import { ProgressController } from '#console/status/progress/controller';
import { setProgressHint, setProgressLabel } from '#console/status/progress/labels';
import { progressPromptError } from '#console/status/progress/prompt';
import { progressCurrent, progressPercentageValue, progressValue } from '#console/status/progress/readback';
import type { ProgressSignalTarget } from '#console/status/progress/terminal';

export type { ProgressSignalTarget } from '#console/status/progress/terminal';

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

	start(): void {
		this.#controller.activate();
	}

	advance(step = 1): void {
		this.#controller.advance(step);
	}

	finish(): void {
		this.#controller.finish();
	}

	fail(): void {
		this.#controller.fail();
	}

	label(value: string): this {
		setProgressLabel(this.#controller, value);

		return this;
	}

	hint(value: string): this {
		setProgressHint(this.#controller, value);

		return this;
	}

	percentage(): number {
		return progressPercentageValue(this.#controller);
	}

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

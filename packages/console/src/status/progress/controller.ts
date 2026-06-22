import { ProgressRenderer } from '#console/status/progress/renderer';
import { ProgressState } from '#console/status/progress/state';
import { ProgressTerminalLifecycle } from '#console/status/progress/terminal';
import type { ProgressSignalTarget } from '#console/status/progress/terminal';

export class ProgressController {
	readonly renderer: ProgressRenderer;
	readonly state: ProgressState;
	readonly terminal: ProgressTerminalLifecycle;

	constructor(total: number, message: string, hint: string, signalTarget: ProgressSignalTarget, handleSignal: () => void) {
		this.state = new ProgressState(total, message, hint);
		this.terminal = new ProgressTerminalLifecycle(signalTarget, handleSignal);
		this.renderer = new ProgressRenderer(this.terminal);
	}

	activate(): void {
		this.state.activate();
		this.render();
	}

	advance(step: number): void {
		this.state.advance(step);
		this.render();
	}

	fail(): void {
		this.state.fail();
		this.render();
		this.terminal.restore();
	}

	finish(): void {
		this.state.finish();
		this.render();
		this.terminal.restore();
	}

	render(): void {
		this.renderer.render(this.state.snapshot());
	}
}

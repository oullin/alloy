import { promptEnvironment } from '#console/environment';
import { renderProgressFrame } from '#console/status/progress/render';
import type { ProgressStateSnapshot } from '#console/status/progress/state';
import type { ProgressTerminalLifecycle } from '#console/status/progress/terminal';

export class ProgressRenderer {
	constructor(private readonly terminal: ProgressTerminalLifecycle) {}

	render(snapshot: ProgressStateSnapshot): void {
		const frame = renderProgressFrame(snapshot);

		this.terminal.beginRender();
		promptEnvironment().output.write(frame);
		this.terminal.commitFrame(frame);
	}
}

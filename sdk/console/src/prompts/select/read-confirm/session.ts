import { eraseRenderedFrame } from '#console/status/frame';
import { renderActiveConfirm, renderCancelledConfirm } from '#console/prompts/select/render-confirm';
import type { ConfirmReadOptions } from '#console/prompts/select/read-confirm/types';

export type ConfirmReaderSession = {
	cancel(): void;
	frame(): string;
	set(value: boolean): void;
	submission(): { frame: string; value: boolean };
	toggle(): void;
	value(): boolean;
};

export const createConfirmReaderSession = (options: ConfirmReadOptions): ConfirmReaderSession => {
	let confirmed = options.default ?? true;
	let frame = renderActiveConfirm(options, confirmed);

	const redraw = (): void => {
		eraseRenderedFrame(frame);
		frame = renderActiveConfirm(options, confirmed);
	};

	return {
		cancel() {
			eraseRenderedFrame(frame);
			renderCancelledConfirm(options, confirmed);
		},
		frame() {
			return frame;
		},
		set(value) {
			confirmed = value;
			redraw();
		},
		submission() {
			return { frame, value: confirmed };
		},
		toggle() {
			confirmed = !confirmed;
			redraw();
		},
		value() {
			return confirmed;
		},
	};
};

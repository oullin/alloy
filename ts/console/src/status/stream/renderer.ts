import { promptEnvironment } from '#console/environment';
import { eraseRenderedFrame } from '#console/status/frame';
import { streamFadeStyles } from '#console/status/stream/fade';
import { renderStreamFrame } from '#console/status/stream/render';
import type { StreamFadeStyle } from '#console/status/stream/fade';

type StreamRenderOptions = {
	fading: string[];
	value: string;
};

export class StreamRenderer {
	readonly #fadeStyles: StreamFadeStyle[];
	#frame = '';

	constructor(fadeStyles = streamFadeStyles()) {
		this.#fadeStyles = fadeStyles;
	}

	current(): string {
		return this.#frame;
	}

	render(options: StreamRenderOptions): void {
		if (this.#frame.length > 0) {
			eraseRenderedFrame(this.#frame);
			this.#frame = '';
		}

		const nextFrame = renderStreamFrame({
			fading: options.fading,
			fadeStyles: this.#fadeStyles,
			value: options.value,
		});

		promptEnvironment().output.write(nextFrame);
		this.#frame = nextFrame;
	}
}

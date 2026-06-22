import type { StreamBuffer } from '#console/status/stream/buffer';
import type { StreamRenderer } from '#console/status/stream/renderer';

type StreamRenderContext = {
	buffer: StreamBuffer;
	renderer: StreamRenderer;
	scheduler: {
		flush(render: () => void): void;
		request(render: () => void): void;
	};
};

export const renderStreamBuffer = ({ buffer, renderer }: StreamRenderContext): void => {
	renderer.render({
		fading: buffer.fading,
		value: buffer.stableValue(),
	});
};

export const requestStreamBufferRender = (context: StreamRenderContext): void => {
	context.scheduler.request(() => {
		renderStreamBuffer(context);
	});
};

export const flushStreamBuffer = (context: StreamRenderContext): void => {
	if (context.buffer.flush()) {
		context.scheduler.flush(() => {
			renderStreamBuffer(context);
		});
	}
};

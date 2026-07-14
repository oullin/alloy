import { StreamBuffer } from '#console/status/stream/buffer';
import { StreamLifecycle } from '#console/status/stream/lifecycle';
import { StreamRenderer } from '#console/status/stream/renderer';
import { StreamRenderScheduler } from '#console/status/stream/scheduler';

export type StreamOutputContext = {
	buffer: StreamBuffer;
	lifecycle: StreamLifecycle;
	renderer: StreamRenderer;
	scheduler: StreamRenderScheduler;
};

export const createStreamOutputContext = (flush: () => void): StreamOutputContext => ({
	buffer: new StreamBuffer(10),
	lifecycle: new StreamLifecycle(flush),
	renderer: new StreamRenderer(),
	scheduler: new StreamRenderScheduler(),
});

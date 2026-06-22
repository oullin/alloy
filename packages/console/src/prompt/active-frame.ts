import { eraseRenderedFrame } from '#console/status/frame';

export type ActivePromptFrame = {
	clear: () => void;
	set: (frame?: string) => void;
};

export const activePromptFrame = (): ActivePromptFrame => {
	let currentFrame: string | undefined;

	return {
		clear: () => {
			if (!currentFrame) {
				return;
			}

			try {
				eraseRenderedFrame(currentFrame);
			} finally {
				currentFrame = undefined;
			}
		},
		set: (frame) => {
			currentFrame = frame;
		},
	};
};

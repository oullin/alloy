export type OpenTuiFrame = {
	text: string;
};

export const renderOpenTuiTextFrame = async (text: string, size: { width?: number; height?: number } = {}): Promise<OpenTuiFrame> => {
	const [{ Text }, { createTestRenderer }] = await Promise.all([import('@opentui/core'), import('@opentui/core/testing')]);

	const { renderer, renderOnce, captureCharFrame } = await createTestRenderer({
		width: size.width ?? 80,
		height: size.height ?? 24,
	});

	try {
		renderer.root.add(Text({ content: text }));

		await renderOnce();

		return { text: captureCharFrame() };
	} finally {
		renderer.destroy();
	}
};

export const pipeStreamSource = async (source: AsyncIterable<string> | Iterable<string>, write: (chunk: string) => void, close: () => void): Promise<void> => {
	try {
		for await (const chunk of source) {
			write(chunk);
		}
	} finally {
		close();
	}
};

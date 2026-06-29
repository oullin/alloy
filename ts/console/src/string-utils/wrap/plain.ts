import { parseStringWidth } from '#console/string-utils/validators/width';
import { visibleWidth } from '#console/string-utils/width';

const wrappedWideWord = (word: string, width: number): string[] => {
	const chunks: string[] = [];

	let chunk = '';
	let chunkWidth = 0;

	for (const char of word) {
		const charWidth = visibleWidth(char);

		if (chunkWidth + charWidth > width) {
			if (chunk.length > 0) {
				chunks.push(chunk);
			}

			chunk = '';
			chunkWidth = 0;
		}

		chunk += char;
		chunkWidth += charWidth;
	}

	return [...chunks, chunk];
};

export const plainWrap = (value: string, width: number): string[] => {
	const parsedWidth = parseStringWidth(width);

	if (parsedWidth === undefined) {
		return [value];
	}

	const lines: string[] = [];

	for (const originalLine of value.split('\n')) {
		let line = '';
		let lineWidth = 0;

		for (const word of originalLine.split(/(\s+)/u)) {
			const wordWidth = visibleWidth(word);

			if (lineWidth + wordWidth <= parsedWidth) {
				line += word;
				lineWidth += wordWidth;
				continue;
			}

			if (line.length > 0) {
				lines.push(line.trimEnd());
				line = '';
				lineWidth = 0;
			}

			if (wordWidth > parsedWidth) {
				const chunks = wrappedWideWord(word, parsedWidth);

				for (const chunk of chunks.slice(0, -1)) {
					lines.push(chunk);
				}

				line = chunks.at(-1) ?? '';
				lineWidth = visibleWidth(line);
				continue;
			}

			line = word.trimStart();
			lineWidth = visibleWidth(line);
		}

		lines.push(line.trimEnd());
	}

	return lines;
};

import { ansiCloseSequence, parseAnsiSegments } from '#console/string-utils/ansi';
import type { AnsiSegment } from '#console/string-utils/ansi';

const codedCharacters = (value: string): AnsiSegment[] => parseAnsiSegments(value).flatMap((segment) => [...segment.text].map((text) => ({ text, codes: segment.codes })));

export const restoreAnsiWrappedLines = (value: string, plainLines: string[]): string[] => {
	const characters = codedCharacters(value);
	const result: string[] = [];

	let characterIndex = 0;

	for (const plainLine of plainLines) {
		let line = '';
		let activeCodes = '';

		for (const plainCharacter of [...plainLine]) {
			while (characterIndex < characters.length && characters[characterIndex]?.text !== plainCharacter) {
				if (characters[characterIndex]?.text === ' ') {
					characterIndex += 1;
					continue;
				}

				break;
			}

			const character = characters[characterIndex];

			if (character === undefined) {
				line += plainCharacter;
				continue;
			}

			if (character.codes !== activeCodes) {
				if (activeCodes !== '') {
					line += '\u001B[0m';
				}

				if (character.codes !== '') {
					line += character.codes;
				}

				activeCodes = character.codes;
			}

			line += plainCharacter;
			characterIndex += 1;
		}

		if (activeCodes !== '') {
			line += ansiCloseSequence(activeCodes);
		}

		result.push(line);
	}

	return result;
};

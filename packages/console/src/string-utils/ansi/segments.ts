import { ansiCode, ansiCodeParts, ESC, isAnsiResetCode, isAnsiStyleSequence } from '#console/string-utils/ansi/codes';

export type AnsiSegment = {
	codes: string;
	text: string;
};

export const parseAnsiSegments = (value: string): AnsiSegment[] => {
	const segments: AnsiSegment[] = [];

	let currentCodes: string[] = [];
	let currentText = '';
	let index = 0;

	while (index < value.length) {
		if (value[index] === ESC && value[index + 1] === '[') {
			if (currentText !== '') {
				segments.push({ text: currentText, codes: activeAnsiCodes(currentCodes) });
				currentText = '';
			}

			const sequence = readAnsiSequence(value, index);

			index += sequence.length;

			if (isAnsiStyleSequence(sequence)) {
				const code = ansiCode(sequence);

				currentCodes = applyAnsiStyleSequence(currentCodes, code, sequence);
			}

			continue;
		}

		currentText += value[index];
		index += 1;
	}

	if (currentText !== '') {
		segments.push({ text: currentText, codes: activeAnsiCodes(currentCodes) });
	}

	return segments;
};

const activeAnsiCodes = (codes: string[]): string => codes.join('');
const hasAnsiStyleCode = (code: string): boolean => ansiCodeParts(code).some((part) => !isAnsiResetCode(part));

const applyAnsiStyleSequence = (codes: string[], code: string, sequence: string): string[] => {
	const next = isAnsiResetCode(code) ? resetAnsiCodes(codes, code) : codes;

	return hasAnsiStyleCode(code) ? [...next, sequence] : next;
};

const resetAnsiCodes = (codes: string[], resetCode: string): string[] => {
	if (ansiCodeParts(resetCode).includes('0')) {
		return [];
	}

	return codes.filter((sequence) => ansiCodeParts(resetCode).every((part) => !ansiCodeMatchesReset(ansiCode(sequence), part)));
};

const ansiCodeMatchesReset = (code: string, resetCode: string): boolean => {
	const parts = ansiCodeParts(code);

	if (resetCode === '22') {
		return parts.some((part) => part === '1' || part === '2');
	}

	if (resetCode === '23') {
		return parts.includes('3');
	}

	if (resetCode === '24') {
		return parts.includes('4');
	}

	if (resetCode === '27') {
		return parts.includes('7');
	}

	if (resetCode === '29') {
		return parts.includes('9');
	}

	if (resetCode === '39') {
		return foregroundAnsiCode(code);
	}

	if (resetCode === '49') {
		return backgroundAnsiCode(code);
	}

	return false;
};

const foregroundAnsiCode = (code: string): boolean => {
	const values = ansiCodeParts(code).map((part) => Number(part));

	return values.includes(38) || values.some((value) => (value >= 30 && value <= 37) || (value >= 90 && value <= 97));
};

const backgroundAnsiCode = (code: string): boolean => {
	const values = ansiCodeParts(code).map((part) => Number(part));

	return values.includes(48) || values.some((value) => (value >= 40 && value <= 47) || (value >= 100 && value <= 107));
};

const readAnsiSequence = (value: string, start: number): string => {
	let sequence = '';
	let index = start;

	while (index < value.length) {
		sequence += value[index];
		index += 1;

		if (isAnsiStyleSequence(sequence)) {
			break;
		}
	}

	return sequence;
};

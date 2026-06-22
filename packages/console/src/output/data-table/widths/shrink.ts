type ShrinkDataTableColumnWidthsOptions = {
	available: number;
	headerWidths: number[];
	natural: number[];
};

export const shrinkDataTableColumnWidths = ({ available, headerWidths, natural }: ShrinkDataTableColumnWidthsOptions): number[] => {
	const totalNatural = natural.reduce((sum, width) => sum + width, 0);

	if (totalNatural <= 0) {
		return natural.map((_, index) => Math.max(headerWidths[index] ?? 0, 1));
	}

	const shrunk = natural.map((width, index) => Math.max(headerWidths[index] ?? 0, Math.floor((available * width) / totalNatural)));

	let remaining = available - shrunk.reduce((sum, width) => sum + width, 0);

	for (const index of dataTableWidthPriority(natural)) {
		if (remaining <= 0) {
			break;
		}

		shrunk[index] += 1;
		remaining -= 1;
	}

	return shrunk;
};

const dataTableWidthPriority = (natural: number[]): number[] =>
	natural
		.map((width, index) => ({ index, width }))
		.sort((left, right) => right.width - left.width)
		.map(({ index }) => index);

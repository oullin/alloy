import type { Amount } from '#money/calculator';

/**
 * Renders minor-unit amounts as display strings using a currency's fraction,
 * separators, symbol (grapheme), and layout template (e.g. `$1`).
 */
export class MoneyFormatter {
	/** Abbreviation steps in major units, ordered largest first. */
	private static readonly SCALES: ReadonlyArray<readonly [bigint, string]> = [
		[1_000_000_000n, 'B'],
		[1_000_000n, 'M'],
		[1_000n, 'K'],
	];

	/** Below this many major units, {@link formatCompact} defers to {@link format}. */
	private static readonly COMPACT_FLOOR_MAJOR_UNITS = 1_000n;

	public constructor(
		public readonly fraction: number,
		public readonly decimal: string,
		public readonly thousand: string,
		public readonly grapheme: string,
		public readonly template: string,
	) {}

	public static create(fraction: number, decimal: string, thousand: string, grapheme: string, template: string): MoneyFormatter {
		return new MoneyFormatter(fraction, decimal, thousand, grapheme, template);
	}

	/** Formats a minor-unit amount as a display string such as `$1,234.56`. */
	public format(amount: Amount): string {
		let raw = this.abs(amount).toString();

		if (raw.length <= this.fraction) {
			raw = '0'.repeat(this.fraction - raw.length + 1) + raw;
		}

		if (this.thousand !== '') {
			for (let index = raw.length - this.fraction - 3; index > 0; index -= 3) {
				raw = `${raw.slice(0, index)}${this.thousand}${raw.slice(index)}`;
			}
		}

		if (this.fraction > 0) {
			raw = `${raw.slice(0, raw.length - this.fraction)}${this.decimal}${raw.slice(raw.length - this.fraction)}`;
		}

		let output = this.template.replace('1', raw).replace('$', this.grapheme);

		if (amount < 0n) {
			output = `-${output}`;
		}

		return output;
	}

	/**
	 * Formats a minor-unit amount abbreviated to a scale suffix, such as
	 * `$1.3M` or `750K`.
	 *
	 * Amounts below one thousand major units are returned at full precision by
	 * {@link format}: rounding `950.00` to `1K` overstates it, and a figure that
	 * short does not need shortening.
	 *
	 * A scale is chosen only when the amount fills at least one of it, largest
	 * first, so a value never renders in a smaller scale than it belongs to. One
	 * decimal is kept only when it carries information — `4B`, not `4.0B` — and
	 * the suffix lands directly after the last digit so it sits inside the
	 * number rather than after a trailing grapheme.
	 *
	 * Everything is computed on the amount's own integer type, so no precision
	 * is lost on the way to the rounded display string.
	 */
	public formatCompact(amount: Amount): string {
		const absolute = this.abs(amount);
		const minorPerMajor = 10n ** BigInt(this.fraction);

		if (absolute < MoneyFormatter.COMPACT_FLOOR_MAJOR_UNITS * minorPerMajor) {
			return this.format(amount);
		}

		for (const [divisor, suffix] of MoneyFormatter.SCALES) {
			const tenths = MoneyFormatter.tenthsOf(absolute, divisor * minorPerMajor);

			if (tenths < 10n) {
				continue;
			}

			return this.renderCompact(amount < 0n ? -tenths : tenths, suffix);
		}

		return this.format(amount);
	}

	/** Converts a minor-unit amount to a floating-point major-unit number. */
	public toMajorUnits(amount: Amount): number {
		if (this.fraction === 0) {
			return Number(amount);
		}

		return Number(amount) / 10 ** this.fraction;
	}

	private abs(amount: Amount): Amount {
		return amount < 0n ? -amount : amount;
	}

	/**
	 * Renders a tenths count at this currency's separators and template.
	 *
	 * The fraction is chosen per value rather than fixed, so a whole number of
	 * tenths asks for no decimals at all. That is what keeps `4B` from becoming
	 * `4.0B` without a second rounding pass.
	 */
	private renderCompact(tenths: Amount, suffix: string): string {
		const isWhole = tenths % 10n === 0n;
		const scaled = new MoneyFormatter(isWhole ? 0 : 1, this.decimal, this.thousand, this.grapheme, this.template);

		return MoneyFormatter.appendSuffix(scaled.format(isWhole ? tenths / 10n : tenths), suffix);
	}

	/** How many tenths of `unit` the amount is, rounded half away from zero. */
	private static tenthsOf(absolute: Amount, unit: Amount): Amount {
		const scaled = absolute * 10n;
		const quotient = scaled / unit;

		return (scaled % unit) * 2n < unit ? quotient : quotient + 1n;
	}

	/**
	 * Places the suffix after the last digit.
	 *
	 * Currencies differ on whether the grapheme leads or trails, so appending to
	 * the end would land the suffix after `.د.إ` for AED while working for USD.
	 * Scanning for an ASCII digit is sufficient: the separators are ASCII and no
	 * grapheme in the dataset contains an ASCII digit.
	 */
	private static appendSuffix(formatted: string, suffix: string): string {
		for (let index = formatted.length - 1; index >= 0; index -= 1) {
			const code = formatted.charCodeAt(index);

			if (code >= 0x30 && code <= 0x39) {
				return `${formatted.slice(0, index + 1)}${suffix}${formatted.slice(index + 1)}`;
			}
		}

		return `${formatted}${suffix}`;
	}
}

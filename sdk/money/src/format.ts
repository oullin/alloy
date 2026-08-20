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

	/**
	 * Formats a minor-unit amount as whole major units, with no decimal part.
	 *
	 * A per-instalment line needs its cents; a headline figure — a reference
	 * price, a due amount stated once on its own line — reads them as noise.
	 * Rounding is half away from zero, so this states a figure rather than
	 * quietly truncating it.
	 *
	 * Computed on the amount's own integer type, so a total past the
	 * float-safe range is stated exactly rather than rounded on the way.
	 */
	public formatWhole(amount: Amount): string {
		const minorPerMajor = 10n ** BigInt(this.fraction);
		const absolute = this.abs(amount);
		const major = MoneyFormatter.divideRounding(absolute, minorPerMajor);
		const whole = new MoneyFormatter(0, this.decimal, this.thousand, this.grapheme, this.template);

		return whole.format(amount < 0n ? -major : major);
	}

	/**
	 * Formats a minor-unit amount abbreviated to a fixed number of significant
	 * digits, such as `$47.2M` or `$4.26M` at three.
	 *
	 * {@link formatCompact} keeps one decimal, which reads well for a single
	 * figure but not for a column: `47.2M` beside `4.2M` gives the second
	 * figure a digit less of information than the first. Fixing the
	 * significant digits instead keeps every row equally precise, which is what
	 * a table of totals wants.
	 *
	 * Trailing zeros are dropped — `4B`, never `4.00B` — so the digit count is
	 * a ceiling rather than padding.
	 */
	public formatCompactSignificant(amount: Amount, significantDigits: number): string {
		const minorPerMajor = 10n ** BigInt(this.fraction);
		const absolute = this.abs(amount);

		if (absolute < MoneyFormatter.COMPACT_FLOOR_MAJOR_UNITS * minorPerMajor) {
			return this.format(amount);
		}

		for (const [divisor, suffix] of MoneyFormatter.SCALES) {
			const unit = divisor * minorPerMajor;

			if (absolute < unit) {
				continue;
			}

			return this.renderSignificant(absolute, unit, significantDigits, amount < 0n, suffix);
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

	/**
	 * Renders `absolute` scaled to `unit` at a fixed significant-digit count.
	 *
	 * The decimal count falls out of how many digits the integer part already
	 * spends: `47` spends two of three, leaving one decimal; `4` spends one,
	 * leaving two. Trailing zeros are then dropped so the count is a ceiling.
	 */
	private renderSignificant(absolute: Amount, unit: Amount, significantDigits: number, negative: boolean, suffix: string): string {
		const integerDigits = Math.max((absolute / unit).toString().length, 1);

		let decimals = Math.max(significantDigits - integerDigits, 0);
		let scaled = MoneyFormatter.divideRounding(absolute * 10n ** BigInt(decimals), unit);

		while (decimals > 0 && scaled % 10n === 0n) {
			scaled /= 10n;
			decimals -= 1;
		}

		const formatter = new MoneyFormatter(decimals, this.decimal, this.thousand, this.grapheme, this.template);

		return MoneyFormatter.appendSuffix(formatter.format(negative ? -scaled : scaled), suffix);
	}

	/** Integer division rounded half away from zero, on non-negative operands. */
	private static divideRounding(value: Amount, unit: Amount): Amount {
		const quotient = value / unit;

		return (value % unit) * 2n < unit ? quotient : quotient + 1n;
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

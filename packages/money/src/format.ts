import type { Amount } from '#money/calculator';

export class MoneyFormatter {
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

	public toMajorUnits(amount: Amount): number {
		if (this.fraction === 0) {
			return Number(amount);
		}

		return Number(amount) / 10 ** this.fraction;
	}

	private abs(amount: Amount): Amount {
		return amount < 0n ? -amount : amount;
	}
}

import { resolveField } from '#workflow/multisteps/response-resolver';

export abstract class Arg {
	public abstract kind(): string;

	public dependsOnJob(): string {
		return '';
	}

	public abstract resolve(vars: Record<string, unknown>, responses: Record<string, unknown>): unknown;
}

export class LiteralArg extends Arg {
	readonly #value: unknown;

	public constructor(value: unknown) {
		super();
		this.#value = value;
	}

	public kind(): string {
		return 'literal';
	}

	public resolve(): unknown {
		return this.#value;
	}
}

export class VariableArg extends Arg {
	public readonly name: string;

	public constructor(name: string) {
		super();
		this.name = name;
	}

	public kind(): string {
		return 'variable';
	}

	public resolve(vars: Record<string, unknown>): unknown {
		if (!Object.hasOwn(vars, this.name)) {
			throw new Error(`variable "${this.name}" not provided`);
		}

		return vars[this.name];
	}
}

export class ResponseArg extends Arg {
	public readonly job: string;
	public readonly field: string;

	public constructor(job: string, field = '') {
		super();
		this.job = job;
		this.field = field;
	}

	public kind(): string {
		return 'response';
	}

	public dependsOnJob(): string {
		return this.job;
	}

	public resolve(_vars: Record<string, unknown>, responses: Record<string, unknown>): unknown {
		if (!Object.hasOwn(responses, this.job)) {
			throw new Error(`response "${this.job}" not available`);
		}

		const raw = responses[this.job];

		if (this.field === '') {
			return raw;
		}

		return resolveField(raw, this.field, this.job);
	}
}

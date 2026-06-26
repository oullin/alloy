import { resolveField } from '#workflow/multisteps/response-resolver';

export class MultiStepResult {
	public readonly responses: Record<string, unknown>;
	public readonly skipped: string[];

	public constructor(responses: Record<string, unknown> = {}, skipped: string[] = []) {
		this.responses = { ...responses };
		this.skipped = [...skipped];
	}

	public as<T>(job: string, field = ''): T {
		const raw = this.responses[job];

		if (!Object.hasOwn(this.responses, job)) {
			throw new Error(`response "${job}" not available`);
		}

		const value = field === '' ? raw : resolveField(raw, field);

		return value as T;
	}
}

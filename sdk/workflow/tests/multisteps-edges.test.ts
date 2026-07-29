import { describe, expect, it } from 'vite-plus/test';

import { AsyncJob, CompileError, LiteralArg, MultiStepEngine, MultiStepResult, MultiStepWorkflow, ResponseArg, SyncJob, WorkflowError } from '@hara/sdk-workflow/multisteps';

describe('multisteps edge cases', () => {
	it('marks skipped dependencies complete enough for downstream jobs', async () => {
		const order: string[] = [];

		const workflow = MultiStepWorkflow.machine(
			'skipped-dependency',
			new SyncJob('start', () => {
				order.push('start');

				return 'start';
			}),
			new SyncJob('optional', () => {
				order.push('optional');

				return 'optional';
			})
				.dependsOn('start')
				.withRunIf(() => false),
			new SyncJob('finish', () => {
				order.push('finish');

				return 'finish';
			}).dependsOn('optional'),
		);

		const result = await new MultiStepEngine().run(workflow);

		expect(order).toEqual(['start', 'finish']);
		expect(result.skipped).toEqual(['optional']);
		expect(result.responses).toEqual({ start: 'start', optional: undefined, finish: 'finish' });
	});

	it('does not run jobs whose run-if dependency response is unresolved', async () => {
		const workflow = MultiStepWorkflow.machine(
			'unresolved-response',
			new SyncJob('optional', () => ({ id: 'x' })).withRunIf(() => false),
			new SyncJob('consumer', () => 'never').withArgs({ id: new ResponseArg('optional', 'id') }),
		);

		await expect(new MultiStepEngine().run(workflow)).rejects.toBeInstanceOf(WorkflowError);
	});

	it('requires jobs and unique non-empty names before running', () => {
		expect(() => MultiStepWorkflow.machine('empty').compile()).toThrow(CompileError);
		expect(() => MultiStepWorkflow.machine('blank', new SyncJob('', () => undefined)).compile()).toThrow(CompileError);
		expect(() => MultiStepWorkflow.machine('dupe', new SyncJob('same', () => 1), new AsyncJob('same', async () => 2)).compile()).toThrow(CompileError);
	});

	it('extracts complete responses and literal arguments after completion', async () => {
		const workflow = MultiStepWorkflow.machine(
			'complete',
			new SyncJob('first', ({ resolved }) => ({ value: resolved.seed })).withArgs({ seed: new LiteralArg(42) }),
			new SyncJob('second', ({ responses }) => ({ first: (responses.first as { value: number }).value })).dependsOn('first'),
		);

		const result = await new MultiStepEngine().run(workflow);

		expect(result).toBeInstanceOf(MultiStepResult);
		expect(result.as<{ value: number }>('first')).toEqual({ value: 42 });
		expect(result.as<number>('second', 'first')).toBe(42);
	});
});

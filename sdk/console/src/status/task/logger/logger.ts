import { createTaskLoggerLimits } from '#console/status/task/logger/limits';
import { createTaskLoggerLabels } from '#console/status/task/logger/labels';
import { notifyTaskLoggerChanged } from '#console/status/task/logger/events';
import { syncTaskLoggerLabel, syncTaskLoggerSubLabel } from '#console/status/task/logger/sync';
import type { TaskLoggerChangeHandler } from '#console/status/task/logger/events';
import type { TaskLoggerLimits } from '#console/status/task/logger/limits';
import type { PartialTaskLogState } from '#console/status/task/logger/lines';
import type { TaskLoggerLabels } from '#console/status/task/logger/labels';
import type { StableTaskMessage } from '#console/status/task/messages';

import { clearTaskLoggerPartial, writeStableTaskLoggerMessage, writeTaskLoggerLine, writeTaskLoggerPartial } from '#console/status/task/logger/methods';

/**
 * Collects a task's live output: a rolling window of log lines plus stable
 * success/warning/error messages, re-rendering the task frame on change.
 */
export class Logger {
	labelValue: string;
	readonly lines: string[] = [];
	readonly stableMessages: StableTaskMessage[] = [];
	subLabelValue: string;
	#partial: PartialTaskLogState = { startIndex: null, value: '' };
	private labels: TaskLoggerLabels;
	private readonly limits: TaskLoggerLimits;

	constructor(
		limit: number,
		label: string,
		subLabel = '',
		private readonly onChange?: TaskLoggerChangeHandler,
	) {
		this.limits = createTaskLoggerLimits(limit);
		this.labels = createTaskLoggerLabels(label, subLabel);
		this.labelValue = this.labels.label;
		this.subLabelValue = this.labels.subLabel;
	}

	get limitValue(): number {
		return this.limits.line;
	}

	/** Appends a line to the rolling log window. */
	line(message: string): void {
		writeTaskLoggerLine(this.lines, message, this.limits.line);
		this.changed();
	}

	/** Alias for {@link line}. */
	log(message: string): void {
		this.line(message);
	}

	/** Replaces the task's main label. */
	label(message: string): void {
		const next = syncTaskLoggerLabel(this.labels, message);

		this.labels = next.labels;
		this.labelValue = next.labelValue;
		this.changed();
	}

	/** Replaces the task's secondary label. */
	subLabel(message: string): void {
		const next = syncTaskLoggerSubLabel(this.labels, message);

		this.labels = next.labels;
		this.subLabelValue = next.subLabelValue;
		this.changed();
	}

	/** Streams a chunk into the current in-progress line without committing it. */
	partial(chunk: string): void {
		this.#partial = writeTaskLoggerPartial(this.lines, this.#partial, chunk, this.limits.line);
		this.changed();
	}

	/** Finalizes the in-progress partial line. */
	commitPartial(): void {
		this.#partial = clearTaskLoggerPartial();
		this.changed();
	}

	/** Alias for {@link line}. */
	info(message: string): void {
		this.line(message);
	}

	/** Records a stable success message that outlives the rolling log window. */
	success(message: string): void {
		this.stable('success', message);
	}

	/** Records a stable warning message that outlives the rolling log window. */
	warning(message: string): void {
		this.stable('warning', message);
	}

	/** Records a stable error message that outlives the rolling log window. */
	error(message: string): void {
		this.stable('error', message);
	}

	private stable(type: StableTaskMessage['type'], message: string): void {
		this.#partial = writeStableTaskLoggerMessage(this.stableMessages, this.lines, type, message, this.limits);
		this.changed();
	}

	private changed(): void {
		notifyTaskLoggerChanged(this.onChange);
	}
}

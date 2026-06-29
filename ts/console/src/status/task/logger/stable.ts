import { parseLogLimit } from '#console/status/validators/limit';
import type { StableTaskMessage } from '#console/status/task/messages';

export const appendStableTaskMessage = (messages: StableTaskMessage[], type: StableTaskMessage['type'], message: string, limit: number): void => {
	const stableLimit = parseLogLimit(limit, 10);

	messages.push({ message, type });

	while (messages.length > stableLimit) {
		messages.shift();
	}
};

import { Stream } from '#console/status/stream/output';

export { Stream } from '#console/status/stream/output';

export function stream(): Stream;

export function stream(source: AsyncIterable<string> | Iterable<string>): Promise<void>;

export function stream(source?: AsyncIterable<string> | Iterable<string>): Stream | Promise<void> {
	const output = new Stream();

	return source === undefined ? output : output.pipe(source);
}

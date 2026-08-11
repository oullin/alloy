import { ERR_INVALID_JSON_UNMARSHAL } from '#money/errors';

export const findTopLevelJsonNumber = (payload: string, property: string): string | null => {
	let index = 0;

	index = skipJsonWhitespace(payload, index);

	if (payload[index] !== '{') {
		throw ERR_INVALID_JSON_UNMARSHAL;
	}

	index += 1;

	while (index < payload.length) {
		index = skipJsonWhitespace(payload, index);

		if (payload[index] === '}') {
			return null;
		}

		if (payload[index] !== '"') {
			throw ERR_INVALID_JSON_UNMARSHAL;
		}

		const key = readJsonString(payload, index);

		index = skipJsonWhitespace(payload, key.end);

		if (payload[index] !== ':') {
			throw ERR_INVALID_JSON_UNMARSHAL;
		}

		index = skipJsonWhitespace(payload, index + 1);

		if (key.value === property) {
			return readJsonNumber(payload, index);
		}

		index = skipJsonValue(payload, index);
		index = skipJsonWhitespace(payload, index);

		if (payload[index] === ',') {
			index += 1;
			continue;
		}

		if (payload[index] === '}') {
			return null;
		}

		throw ERR_INVALID_JSON_UNMARSHAL;
	}

	throw ERR_INVALID_JSON_UNMARSHAL;
};

const JSON_NUMBER = /-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?/u;

/** The same grammar, anchored, so a quoted amount must be a number and nothing else. */
const QUOTED_JSON_NUMBER = /^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?$/u;

const skipJsonWhitespace = (payload: string, start: number): number => {
	let index = start;

	while (index < payload.length && /\s/u.test(payload[index] ?? '')) {
		index += 1;
	}

	return index;
};

const readJsonString = (payload: string, start: number): { value: string; end: number } => {
	let index = start + 1;
	let value = '';

	while (index < payload.length) {
		const character = payload[index];

		if (character === '"') {
			return { value, end: index + 1 };
		}

		if (character === '\\') {
			const escaped = payload[index + 1];

			if (escaped === undefined) {
				throw ERR_INVALID_JSON_UNMARSHAL;
			}

			value += `\\${escaped}`;
			index += 2;
			continue;
		}

		value += character;
		index += 1;
	}

	throw ERR_INVALID_JSON_UNMARSHAL;
};

const readJsonNumber = (payload: string, start: number): string => {
	// A quoted amount is as valid as a bare one. Go types the field as
	// json.Number, which encoding/json fills from a JSON string as readily as
	// from a JSON number, and `MoneyValue.toJSON()` emits the quoted form so an
	// int64 survives a JavaScript consumer's `JSON.parse`. Reading only the bare
	// form left this runtime stricter than its twin and made
	// `unmarshal(JSON.stringify(value))` throw on its own output.
	//
	// The grammar is anchored rather than merely prefix-matched: Go rejects
	// `" 125 "` and `"12abc"`, so accepting a numeric prefix here would trade one
	// divergence for another.
	if (payload[start] === '"') {
		const quoted = readJsonString(payload, start);

		if (!QUOTED_JSON_NUMBER.test(quoted.value)) {
			throw ERR_INVALID_JSON_UNMARSHAL;
		}

		return quoted.value;
	}

	const match = JSON_NUMBER.exec(payload.slice(start));

	if (match?.index !== 0) {
		throw ERR_INVALID_JSON_UNMARSHAL;
	}

	return match[0];
};

const skipJsonValue = (payload: string, start: number): number => {
	const character = payload[start];

	if (character === '"') {
		return readJsonString(payload, start).end;
	}

	if (character === '{') {
		return skipJsonComposite(payload, start, '{', '}');
	}

	if (character === '[') {
		return skipJsonComposite(payload, start, '[', ']');
	}

	const number = JSON_NUMBER.exec(payload.slice(start));

	if (number?.index === 0) {
		return start + number[0].length;
	}

	for (const literal of ['true', 'false', 'null']) {
		if (payload.startsWith(literal, start)) {
			return start + literal.length;
		}
	}

	throw ERR_INVALID_JSON_UNMARSHAL;
};

const skipJsonComposite = (payload: string, start: number, open: string, close: string): number => {
	let depth = 0;
	let index = start;

	while (index < payload.length) {
		const character = payload[index];

		if (character === '"') {
			index = readJsonString(payload, index).end;
			continue;
		}

		if (character === open) {
			depth += 1;
		} else if (character === close) {
			depth -= 1;

			if (depth === 0) {
				return index + 1;
			}
		}

		index += 1;
	}

	throw ERR_INVALID_JSON_UNMARSHAL;
};

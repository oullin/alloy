export class MoneyError extends Error {
	public constructor(
		public readonly code: string,
		message: string,
		public readonly cause?: unknown,
	) {
		super(message);
		this.name = 'MoneyError';
	}
}

const error = (code: string, message: string): MoneyError => new MoneyError(code, message);

export const ERR_CURRENCY_MISMATCH = error('ERR_CURRENCY_MISMATCH', "currencies don't match");

export const ERR_CURRENCY_NOT_FOUND = error('ERR_CURRENCY_NOT_FOUND', 'currency not found');

export const ERR_CURRENCY_CONVERSION_NOT_FOUND = error('ERR_CURRENCY_CONVERSION_NOT_FOUND', 'currency conversion rate not found');

export const ERR_NO_CURRENCY_INSTANCE = error('ERR_NO_CURRENCY_INSTANCE', 'money instance has no currency');

export const ERR_NO_CURRENCY_MANAGER = error('ERR_NO_CURRENCY_MANAGER', 'currency manager cannot be nil');

export const ERR_NO_CURRENCY_MAP_DATASET = error('ERR_NO_CURRENCY_MAP_DATASET', 'currency map dataset cannot be nil or empty');

export const ERR_INVALID_JSON_UNMARSHAL = error('ERR_INVALID_JSON_UNMARSHAL', 'invalid json unmarshal');

export const ERR_INVALID_EXCHANGE_RATE = error('ERR_INVALID_EXCHANGE_RATE', 'invalid exchange rate');

export const ERR_NO_JSON_PARSER_PROVIDED = error('ERR_NO_JSON_PARSER_PROVIDED', 'no json parser provided');

export const ERR_JSON_UNMARSHAL_FUNC_NIL = error('ERR_JSON_UNMARSHAL_FUNC_NIL', 'money.JSON: unmarshal function cannot be nil');

export const ERR_JSON_MARSHAL_FUNC_NIL = error('ERR_JSON_MARSHAL_FUNC_NIL', 'money.JSON: marshal function cannot be nil');

export const ERR_NO_MONEY_PROVIDED = error('ERR_NO_MONEY_PROVIDED', 'no money objects provided');

export const ERR_INVALID_MONEY_STRING = error('ERR_INVALID_MONEY_STRING', 'invalid money string format');

export const ERR_CURRENCY_NOT_SPECIFIED = error('ERR_CURRENCY_NOT_SPECIFIED', 'currency not specified or detected');

export const ERR_NO_MULTIPLIERS_PROVIDED = error('ERR_NO_MULTIPLIERS_PROVIDED', 'no multipliers provided');

export const ERR_NO_CONVERTER_PROVIDED = error('ERR_NO_CONVERTER_PROVIDED', 'no converter provided');

export const ERR_EMPTY_AMOUNT_STRING = error('ERR_EMPTY_AMOUNT_STRING', 'amount string cannot be empty');

export const ERR_INVALID_AMOUNT_MULTIPLE = error('ERR_INVALID_AMOUNT_MULTIPLE', 'invalid amount: multiple decimal points');

export const ERR_INVALID_AMOUNT_FRACTION = error('ERR_INVALID_AMOUNT_FRACTION', 'too many decimal places for curr');

export const ERR_INVALID_AMOUNT = error('ERR_INVALID_AMOUNT', 'invalid amount');

export const ERR_INVALID_SPLIT = error('ERR_INVALID_SPLIT', 'split must be higher than zero');

export const ERR_NO_RATIOS_PROVIDED = error('ERR_NO_RATIOS_PROVIDED', 'no ratios specified');

export const ERR_NEGATIVE_RATIOS = error('ERR_NEGATIVE_RATIOS', 'negative ratios not allowed');

export const ERR_RATIOS_EXCEED_MAX_INT = error('ERR_RATIOS_EXCEED_MAX_INT', 'sum of given ratios exceeds max int');

export const ERR_INVALID_AGGREGATOR_PROVIDER = error('ERR_INVALID_AGGREGATOR_PROVIDER', 'invalid aggregator: nil manager');

export const ERR_OVERFLOW = error('ERR_OVERFLOW', 'arithmetic operation resulted in overflow');

export const ERR_PARSER_NOT_PROVIDED = error('ERR_PARSER_NOT_PROVIDED', 'parser was not provided');

export const ERR_PARSER_INVALID_STATE = error('ERR_PARSER_INVALID_STATE', 'parser is nil or iso is nil');

export const MoneyErrors = Object.freeze({
	ERR_CURRENCY_MISMATCH,
	ERR_CURRENCY_NOT_FOUND,
	ERR_CURRENCY_CONVERSION_NOT_FOUND,
	ERR_NO_CURRENCY_INSTANCE,
	ERR_NO_CURRENCY_MANAGER,
	ERR_NO_CURRENCY_MAP_DATASET,
	ERR_INVALID_JSON_UNMARSHAL,
	ERR_INVALID_EXCHANGE_RATE,
	ERR_NO_JSON_PARSER_PROVIDED,
	ERR_JSON_UNMARSHAL_FUNC_NIL,
	ERR_JSON_MARSHAL_FUNC_NIL,
	ERR_NO_MONEY_PROVIDED,
	ERR_INVALID_MONEY_STRING,
	ERR_CURRENCY_NOT_SPECIFIED,
	ERR_NO_MULTIPLIERS_PROVIDED,
	ERR_NO_CONVERTER_PROVIDED,
	ERR_EMPTY_AMOUNT_STRING,
	ERR_INVALID_AMOUNT_MULTIPLE,
	ERR_INVALID_AMOUNT_FRACTION,
	ERR_INVALID_AMOUNT,
	ERR_INVALID_SPLIT,
	ERR_NO_RATIOS_PROVIDED,
	ERR_NEGATIVE_RATIOS,
	ERR_RATIOS_EXCEED_MAX_INT,
	ERR_INVALID_AGGREGATOR_PROVIDER,
	ERR_OVERFLOW,
	ERR_PARSER_NOT_PROVIDED,
	ERR_PARSER_INVALID_STATE,
});

export const allMoneyErrors = (): MoneyError[] => Object.values(MoneyErrors);

export const invalidJsonUnmarshalFrom = (cause: unknown): MoneyError => new MoneyError('ERR_INVALID_JSON_UNMARSHAL', 'invalid json unmarshal', cause);

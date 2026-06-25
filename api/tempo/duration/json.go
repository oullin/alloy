package duration

import "strconv"

func (value *Span) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(value.ISOString())), nil
}

func (value *Span) UnmarshalJSON(data []byte) error {
	input, err := strconv.Unquote(string(data))

	if err != nil {
		return err
	}

	parsed, err := Parse(input)

	if err != nil {
		return err
	}

	*value = parsed

	return nil
}

package duration

import "strconv"

func (duration Duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(duration.ISOString())), nil
}

func (duration *Duration) UnmarshalJSON(data []byte) error {
	input, err := strconv.Unquote(string(data))

	if err != nil {
		return err
	}

	parsed, err := Parse(input)

	if err != nil {
		return err
	}

	*duration = parsed

	return nil
}

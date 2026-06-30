package fortify

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

const maxRequestBody = 1 << 20

func readInput(r *http.Request) (map[string]any, error) {
	contentType := r.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(contentType)

	if mediaType == "application/json" {
		return readJSONInput(r)
	}

	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	input := make(map[string]any, len(r.Form))

	for key, values := range r.Form {
		if len(values) == 1 {
			input[key] = values[0]

			continue
		}

		input[key] = values
	}

	return input, nil
}

func readJSONInput(r *http.Request) (map[string]any, error) {
	data, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody))

	if err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]any{}, nil
	}

	var input map[string]any

	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	if input == nil {
		input = map[string]any{}
	}

	return input, nil
}

func stringInput(input map[string]any, key string) string {
	value, ok := input[key]

	if !ok || value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func boolInput(input map[string]any, key string) bool {
	value, ok := input[key]

	if !ok || value == nil {
		return false
	}

	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		v := strings.TrimSpace(strings.ToLower(typed))

		return v == "1" || v == "true" || v == "on" || v == "yes"
	case float64:
		return typed != 0
	default:
		return false
	}
}

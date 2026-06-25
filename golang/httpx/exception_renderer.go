package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ValidationError carries field-level validation failures for JSON APIs.
type ValidationError struct {
	Message string
	Errors  map[string][]string
}

// ExceptionRenderer renders errors into consistent headless JSON responses.
type ExceptionRenderer struct {
	Debug bool
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "validation failed"
}

// Render writes a JSON error response for err.
func (r ExceptionRenderer) Render(w http.ResponseWriter, err error) {
	status, message, headers, payload := r.resolve(err)

	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payloadFor(status, message, payload, r.Debug, err))
}

func (r ExceptionRenderer) resolve(err error) (int, string, http.Header, any) {
	if err == nil {
		return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), http.Header{}, nil
	}

	var validation *ValidationError

	if errors.As(err, &validation) {
		return http.StatusUnprocessableEntity, validation.Error(), http.Header{}, map[string]any{
			"errors": validation.Errors,
		}
	}

	var response *HttpResponseError

	if errors.As(err, &response) {
		status := response.StatusCode

		if status == 0 {
			status = http.StatusInternalServerError
		}

		message := response.Message

		if message == "" {
			message = http.StatusText(status)
		}

		return status, message, response.Headers, response.Response
	}

	return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), http.Header{}, nil
}

func payloadFor(status int, message string, payload any, debug bool, err error) map[string]any {
	body := map[string]any{
		"message": message,
		"status":  status,
	}

	if values, ok := payload.(map[string]any); ok {
		for key, value := range values {
			body[key] = value
		}
	} else if payload != nil {
		body["error"] = payload
	}

	if debug && err != nil {
		body["debug"] = err.Error()
	}

	return body
}

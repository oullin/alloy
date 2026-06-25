package access

import (
	"encoding/json"
	"net/http"
)

// ModelResolver resolves the model argument for a route authorization check.
type ModelResolver func(r *http.Request) (any, error)

// AuthorizeMiddleware checks an ability before the route handler runs.
func AuthorizeMiddleware(gate *Gate, ability string, model any) func(http.Handler) http.Handler {
	return AuthorizeResolvedMiddleware(gate, ability, func(*http.Request) (any, error) {
		return model, nil
	})
}

// AuthorizeResolvedMiddleware checks an ability with a request-resolved model.
func AuthorizeResolvedMiddleware(gate *Gate, ability string, resolve ModelResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			model, err := resolve(r)

			if err != nil {
				writeDenial(w, DenyAsNotFound("resource not found"))

				return
			}

			response := gate.Inspect(r.Context(), ability, model)

			if !response.Allowed {
				writeDenial(w, response)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeDenial(w http.ResponseWriter, response Response) {
	status := response.StatusCode

	if status == 0 {
		status = http.StatusForbidden
	}

	message := response.Message

	if message == "" {
		message = http.StatusText(status)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": message,
		"status":  status,
	})
}

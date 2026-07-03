package fortify

import (
	"net/http"

	cauth "github.com/oullin/alloy/packages/foundation/contracts/auth"
)

// RegisterConfig controls registration handler behavior.
type RegisterConfig struct {
	AutoLogin bool
}

// RegisterInput is the normalized input passed to application registration code.
type RegisterInput struct {
	Name     string
	Email    string
	Password string
	Fields   map[string]any
}

// NewRegisterHandler creates a headless registration endpoint.
func NewRegisterHandler(create RegisterUser, guard cauth.StatefulGuard, cfg RegisterConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := readInput(r)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")

			return
		}

		email := stringInput(input, "email")
		password := stringInput(input, "password")

		if email == "" {
			writeValidation(w, "email", "required")

			return
		}

		if password == "" {
			writeValidation(w, "password", "required")

			return
		}

		if confirmation := stringInput(input, "password_confirmation"); confirmation != "" && confirmation != password {
			writeValidation(w, "password_confirmation", "does not match")

			return
		}

		user, err := create(r.Context(), RegisterInput{
			Name:     stringInput(input, "name"),
			Email:    email,
			Password: password,
			Fields:   input,
		})

		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "registration failed")

			return
		}

		if cfg.AutoLogin && guard != nil {
			if err := guard.Login(r.Context(), user, false); err != nil {
				writeError(w, http.StatusInternalServerError, "login failed")

				return
			}
		}

		writeJSON(w, http.StatusCreated, jsonMessage{OK: true, Message: "registered"})
	}
}

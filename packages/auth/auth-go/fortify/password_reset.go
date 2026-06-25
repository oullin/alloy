package fortify

import (
	"errors"
	"net/http"

	"github.com/oullin/alloy/auth/passwords"
)

// NewForgotPasswordHandler sends a reset link with an enumeration-safe response.
func NewForgotPasswordHandler(sender ResetLinkSender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := readInput(r)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")

			return
		}

		email := stringInput(input, "email")

		if email == "" {
			writeValidation(w, "email", "required")

			return
		}

		err = sender.SendResetLink(r.Context(), email)

		if errors.Is(err, passwords.ErrResetLinkThrottled) {
			writeError(w, http.StatusTooManyRequests, "too many reset attempts")

			return
		}

		writeOK(w, "password reset link sent")
	}
}

// NewResetPasswordHandler resets a password after token validation succeeds.
func NewResetPasswordHandler(resetter PasswordResetter, resetFn passwords.ResetCallback) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := readInput(r)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")

			return
		}

		email := stringInput(input, "email")
		token := stringInput(input, "token")
		password := stringInput(input, "password")

		if email == "" {
			writeValidation(w, "email", "required")

			return
		}

		if token == "" {
			writeValidation(w, "token", "required")

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

		err = resetter.Reset(r.Context(), map[string]any{
			"email":    email,
			"token":    token,
			"password": password,
		}, resetFn)

		if err != nil {
			writeValidation(w, "token", "invalid or expired")

			return
		}

		writeOK(w, "password reset")
	}
}

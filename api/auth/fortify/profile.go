package fortify

import (
	"net/http"

	cauth "github.com/oullin/alloy/api/contracts/auth"
)

// NewUpdateProfileHandler delegates profile persistence to the application.
func NewUpdateProfileHandler(guard cauth.Guard, update ProfileUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := guard.User(r.Context())

		if err != nil || user == nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")

			return
		}

		input, err := readInput(r)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")

			return
		}

		if update == nil {
			writeError(w, http.StatusInternalServerError, "profile update is not configured")

			return
		}

		if err := update(r.Context(), user, input); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "profile update failed")

			return
		}

		writeOK(w, "profile updated")
	}
}

// NewUpdatePasswordHandler verifies the current password and persists a new hash.
func NewUpdatePasswordHandler(guard cauth.Guard, hasher cauth.PasswordHasher, update PasswordUpdater, invalidators ...PasswordSessionInvalidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := guard.User(r.Context())

		if err != nil || user == nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")

			return
		}

		input, err := readInput(r)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid request")

			return
		}

		currentPassword := stringInput(input, "current_password")
		password := stringInput(input, "password")

		if currentPassword == "" {
			writeValidation(w, "current_password", "required")

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

		ok, err := hasher.Check(r.Context(), currentPassword, user.GetAuthPassword())

		if err != nil || !ok {
			writeValidation(w, "current_password", "invalid password")

			return
		}

		hashed, err := hasher.Hash(r.Context(), password)

		if err != nil {
			writeError(w, http.StatusInternalServerError, "password update failed")

			return
		}

		if update != nil {
			if err := update(r.Context(), user, hashed); err != nil {
				writeError(w, http.StatusUnprocessableEntity, "password update failed")

				return
			}
		}

		user.SetAuthPassword(hashed)

		for _, invalidate := range invalidators {
			if invalidate == nil {
				continue
			}

			if err := invalidate(r.Context(), user); err != nil {
				writeError(w, http.StatusInternalServerError, "session invalidation failed")

				return
			}
		}

		writeOK(w, "password updated")
	}
}

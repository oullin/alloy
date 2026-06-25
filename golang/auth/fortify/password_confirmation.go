package fortify

import (
	"net/http"
	"time"

	cauth "github.com/oullin/alloy/auth/contracts/auth"
)

// NewConfirmPasswordHandler records a recent password confirmation.
func NewConfirmPasswordHandler(guard cauth.Guard, hasher cauth.PasswordHasher, session PasswordConfirmationSession) http.HandlerFunc {
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

		password := stringInput(input, "password")

		if password == "" {
			writeValidation(w, "password", "required")

			return
		}

		ok, err := hasher.Check(r.Context(), password, user.GetAuthPassword())

		if err != nil || !ok {
			writeValidation(w, "password", "invalid password")

			return
		}

		session.Put(PasswordConfirmedAtKey, time.Now().Unix())
		writeOK(w, "password confirmed")
	}
}

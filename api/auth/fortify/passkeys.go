package fortify

import (
	"net/http"

	cauth "github.com/oullin/alloy/api/contracts/auth"
	"github.com/oullin/alloy/auth/passkeys"
)

// NewBeginPasskeyRegistrationHandler returns WebAuthn registration options.
func NewBeginPasskeyRegistrationHandler(guard cauth.Guard, service *passkeys.Service, key PasskeySessionKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
			return
		}

		options, err := service.BeginRegistration(r.Context(), passkeySessionKey(r, key), user)

		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "passkey registration failed")

			return
		}

		writeJSON(w, http.StatusOK, options)
	}
}

// NewFinishPasskeyRegistrationHandler validates and stores a passkey.
func NewFinishPasskeyRegistrationHandler(guard cauth.Guard, service *passkeys.Service, key PasskeySessionKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := authenticatedUser(w, r, guard)

		if !ok {
			return
		}

		if _, err := service.FinishRegistration(r.Context(), passkeySessionKey(r, key), user, r); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "invalid passkey registration")

			return
		}

		writeOK(w, "passkey registered")
	}
}

// NewBeginPasskeyLoginHandler returns discoverable passkey login options.
func NewBeginPasskeyLoginHandler(service *passkeys.Service, key PasskeySessionKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		options, err := service.BeginDiscoverableLogin(r.Context(), passkeySessionKey(r, key))

		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "passkey login failed")

			return
		}

		writeJSON(w, http.StatusOK, options)
	}
}

// NewFinishPasskeyLoginHandler validates a passkey login and logs the user in.
func NewFinishPasskeyLoginHandler(guard cauth.StatefulGuard, service *passkeys.Service, key PasskeySessionKey, resolve PasskeyUserResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _, err := service.FinishPasskeyLogin(r.Context(), passkeySessionKey(r, key), r, resolve)

		if err != nil || user == nil {
			writeError(w, http.StatusUnprocessableEntity, "invalid passkey login")

			return
		}

		if err := guard.Login(r.Context(), user, false); err != nil {
			writeError(w, http.StatusInternalServerError, "login failed")

			return
		}

		writeOK(w, "authenticated")
	}
}

func passkeySessionKey(r *http.Request, key PasskeySessionKey) string {
	if key != nil {
		return key(r)
	}

	if header := r.Header.Get("X-WebAuthn-Session"); header != "" {
		return header
	}

	return "default"
}

package fortify

import (
	"net/http"

	cauth "alloy.dev/backend/contracts/auth"
)

// NewVerificationNotificationHandler sends the current user a verification link.
func NewVerificationNotificationHandler(guard cauth.Guard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := guard.User(r.Context())

		if err != nil || user == nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")

			return
		}

		verifiable, ok := user.(cauth.MustVerifyEmail)

		if !ok || verifiable.HasVerifiedEmail() {
			writeOK(w, "already verified")

			return
		}

		sender, ok := user.(cauth.EmailVerificationNotificationSender)

		if !ok {
			writeError(w, http.StatusUnprocessableEntity, "email verification is not configured")

			return
		}

		sender.SendEmailVerificationNotification(r.Context())
		writeOK(w, "verification link sent")
	}
}

// NewVerifyEmailHandler delegates signed-link verification to the application.
func NewVerifyEmailHandler(guard cauth.Guard, verify VerifyEmail) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := guard.User(r.Context())

		if err != nil || user == nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")

			return
		}

		verifiable, ok := user.(cauth.MustVerifyEmail)

		if !ok {
			writeError(w, http.StatusUnprocessableEntity, "user cannot verify email")

			return
		}

		if verifiable.HasVerifiedEmail() {
			writeOK(w, "already verified")

			return
		}

		if verify == nil {
			writeError(w, http.StatusInternalServerError, "email verification is not configured")

			return
		}

		if err := verify(r.Context(), r, verifiable); err != nil {
			writeError(w, http.StatusForbidden, "invalid verification link")

			return
		}

		writeOK(w, "email verified")
	}
}
